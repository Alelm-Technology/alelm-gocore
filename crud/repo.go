package crud

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/alelmtech/gocore/pagination"
	"github.com/alelmtech/gocore/sqlext"
)

type TableInfo struct {
	Name       string
	Columns    []string
	UpdColumns []string
	OrderBy    string
	// SearchCols are columns scanned with ILIKE for the Search term.
	SearchCols []string
	// SortCols maps a client sort key to a safe SQL column. When empty, the
	// generic List also allows sorting by any real table column (whitelisted).
	SortCols map[string]string
	// FilterCols declares server-side equality / date-range filters.
	FilterCols []FilterCol
	// ExtraScope optionally injects a caller-derived WHERE fragment for
	// row-level data scoping (e.g. wali/teacher viewers). It receives the
	// request context and the next available ($1-based) parameter index, and
	// returns a fragment referencing $nextParam.. and its arguments. Return
	// ("", nil) to apply no extra scope (e.g. for tenant admins).
	ExtraScope func(ctx context.Context, nextParam int) (string, []interface{})
	// ValidateCreate optionally validates that the entity being created falls
	// within the caller's scope (used to prevent a scoped writer, e.g. a
	// teacher, from creating rows for resources they do not own). It is invoked
	// before the INSERT. Return nil to allow the create; return an error
	// (e.g. sql.ErrNoRows) to reject it.
	ValidateCreate func(ctx context.Context, entity interface{}) error
	// ValidateUpdate optionally validates the entity being updated BEFORE the
	// UPDATE is run. This is the write-side counterpart to ValidateCreate: it
	// closes the PUT-reassignment gap where a scoped writer could change an
	// ownership FK (e.g. santri_id) to a resource outside their scope even
	// though the existing row was in scope. It runs after overlayExisting so the
	// entity's FK columns reflect the full (post-overlay) update body. Return
	// nil to allow the update; return an error (e.g. sql.ErrNoRows) to reject
	// it before any rows are changed.
	ValidateUpdate func(ctx context.Context, entity interface{}) error
}

// FilterCol maps a client filter key (f_<key> query param) to a SQL column.
type FilterCol struct {
	Key  string // client filter key (without f_ prefix)
	Col  string // safe SQL column expression
	Type string // "eq" (default) or "date_range" (value "from,to")
}

type ScopeConfig struct {
	Column string
	Source func(ctx context.Context) interface{}
}

type Repository[E any, ID comparable] interface {
	Create(ctx context.Context, entity *E) error
	FindByID(ctx context.Context, id ID) (*E, error)
	Update(ctx context.Context, entity *E) error
	Delete(ctx context.Context, id ID) error
	List(ctx context.Context, page pagination.Pagination) ([]E, int, error)
}

// actorCtxKey carries the current audit actor (usually the logged-in user id)
// for auto-populating created_by / updated_by columns.
type actorCtxKey struct{}

// WithActor returns ctx carrying actorID (e.g. the logged-in user's uuid-as-string).
// BaseRepo auto-writes it into created_by on INSERT and updated_by on UPDATE
// whenever the entity has those db-tagged fields. An absent or empty actor is
// written as NULL, so system / seed writers are recorded as creator-less rows.
func WithActor(ctx context.Context, actorID string) context.Context {
	return context.WithValue(ctx, actorCtxKey{}, actorID)
}

// Actor returns the audit actor stored via WithActor, or "" when absent.
func Actor(ctx context.Context) string {
	s, _ := ctx.Value(actorCtxKey{}).(string)
	return s
}

// ActorArg yields the actor value to bind into a hand-written SQL audit
// column: the actor string when present, else SQL NULL (nullable audit columns
// stay NULL for system writes).
func ActorArg(ctx context.Context) interface{} {
	if a := Actor(ctx); a != "" {
		return a
	}
	return nil
}

// actorArg is the deprecated unexported alias, kept for internal callers.
func actorArg(ctx context.Context) interface{} { return ActorArg(ctx) }

type fieldInfo struct {
	idx  int
	name string
}

var typeFields sync.Map

func cachedFields(t reflect.Type) []fieldInfo {
	if cached, ok := typeFields.Load(t); ok {
		return cached.([]fieldInfo)
	}
	var fields []fieldInfo
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("db")
		if tag == "" || tag == "-" {
			continue
		}
		if idx := strings.IndexByte(tag, ','); idx != -1 {
			tag = tag[:idx]
		}
		fields = append(fields, fieldInfo{idx: i, name: tag})
	}
	typeFields.Store(t, fields)
	return fields
}

type BaseRepo[E any, ID comparable] struct {
	db    *sqlx.DB
	table TableInfo
	scope *ScopeConfig

	q struct {
		create        string
		createCols    []string
		updCols       []string
		findByID      string
		update        string
		updateNoActor string
		delete        string
		listCnt       string
		listData      string
	}
	qOnce sync.Once

	fields []fieldInfo
	fOnce  sync.Once

	hasTS    bool // struct has created_at && updated_at
	hasActor bool // struct has created_by
	hasUpdBy bool // struct has updated_by
}

func NewBaseRepo[E any, ID comparable](db *sqlx.DB, table TableInfo, scope *ScopeConfig) *BaseRepo[E, ID] {
	return &BaseRepo[E, ID]{db: db, table: table, scope: scope}
}

func (r *BaseRepo[E, ID]) initType() {
	r.fOnce.Do(func() {
		var e E
		t := reflect.TypeOf(e)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		r.fields = cachedFields(t)
	})
}

func (r *BaseRepo[E, ID]) hasField(name string) bool {
	r.initType()
	for _, f := range r.fields {
		if f.name == name {
			return true
		}
	}
	return false
}

func (r *BaseRepo[E, ID]) init() {
	r.qOnce.Do(func() {
		hasScope := r.scope != nil
		hasTS := r.hasField("created_at") && r.hasField("updated_at")
		hasActor := r.hasField("created_by")
		hasUpdBy := r.hasField("updated_by")
		r.hasTS, r.hasActor, r.hasUpdBy = hasTS, hasActor, hasUpdBy

		managed := map[string]bool{}
		if hasTS {
			managed["created_at"] = true
			managed["updated_at"] = true
		}
		if hasActor {
			managed["created_by"] = true
		}

		{
			cols := []string{"id"}
			params := []string{"$1"}
			ph := 1
			if hasScope {
				ph++
				cols = append(cols, r.scope.Column)
				params = append(params, fmt.Sprintf("$%d", ph))
			}
			createCols := make([]string, 0, len(r.table.Columns))
			for _, c := range r.table.Columns {
				if hasScope && c == r.scope.Column {
					continue
				}
				if managed[c] {
					continue
				}
				createCols = append(createCols, c)
			}
			for _, c := range createCols {
				ph++
				cols = append(cols, c)
				params = append(params, fmt.Sprintf("$%d", ph))
			}
			if hasTS {
				cols = append(cols, "created_at", "updated_at")
				params = append(params, "NOW()", "NOW()")
			}
			if hasActor {
				ph++
				cols = append(cols, "created_by")
				params = append(params, fmt.Sprintf("$%d", ph))
			}
			r.q.createCols = createCols
			r.q.create = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				r.table.Name, strings.Join(cols, ", "), strings.Join(params, ", "))
		}

		if hasScope {
			r.q.findByID = fmt.Sprintf("SELECT * FROM %s WHERE id=$1 AND %s=$2", r.table.Name, r.scope.Column)
			r.q.delete = fmt.Sprintf("DELETE FROM %s WHERE id=$1 AND %s=$2", r.table.Name, r.scope.Column)
		} else {
			r.q.findByID = fmt.Sprintf("SELECT * FROM %s WHERE id=$1", r.table.Name)
			r.q.delete = fmt.Sprintf("DELETE FROM %s WHERE id=$1", r.table.Name)
		}

		{
			paramIdx := 1
			updCols := make([]string, 0, len(r.table.UpdColumns))
			for _, c := range r.table.UpdColumns {
				if hasScope && c == r.scope.Column {
					continue
				}
				if hasTS && c == "updated_at" {
					continue
				}
				if hasTS && c == "created_at" {
					continue
				}
				if hasUpdBy && c == "updated_by" {
					continue
				}
				updCols = append(updCols, c)
			}
			bizParts := make([]string, 0, len(updCols)+2)
			for _, c := range updCols {
				bizParts = append(bizParts, fmt.Sprintf("%s=$%d", c, paramIdx))
				paramIdx++
			}
			last := paramIdx - 1
			r.q.updCols = updCols

			buildUpdate := func(setParts []string, idIdx int) string {
				if hasScope {
					return fmt.Sprintf("UPDATE %s SET %s WHERE id=$%d AND %s=$%d",
						r.table.Name, strings.Join(setParts, ", "), idIdx, r.scope.Column, idIdx+1)
				}
				return fmt.Sprintf("UPDATE %s SET %s WHERE id=$%d",
					r.table.Name, strings.Join(setParts, ", "), idIdx)
			}
			withTS := func(parts []string) []string {
				if hasTS {
					return append(parts, "updated_at=NOW()")
				}
				return parts
			}

			if hasUpdBy {
				withActor := append([]string{}, bizParts...)
				withActor = append(withActor, fmt.Sprintf("updated_by=$%d", last+1))
				r.q.update = buildUpdate(withTS(withActor), last+2)

				noActor := append([]string{}, bizParts...)
				r.q.updateNoActor = buildUpdate(withTS(noActor), last+1)
			} else {
				r.q.update = buildUpdate(withTS(append([]string{}, bizParts...)), last+1)
				r.q.updateNoActor = r.q.update
			}
		}

		if hasScope {
			r.q.listCnt = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s=$1", r.table.Name, r.scope.Column)
			r.q.listData = fmt.Sprintf("SELECT * FROM %s WHERE %s=$1 ORDER BY %s LIMIT $2 OFFSET $3",
				r.table.Name, r.scope.Column, r.table.OrderBy)
		} else {
			r.q.listCnt = fmt.Sprintf("SELECT COUNT(*) FROM %s", r.table.Name)
			r.q.listData = fmt.Sprintf("SELECT * FROM %s ORDER BY %s LIMIT $1 OFFSET $2",
				r.table.Name, r.table.OrderBy)
		}
	})
}

func (r *BaseRepo[E, ID]) fieldVal(entity *E, name string) interface{} {
	if entity == nil {
		return nil
	}
	v := reflect.ValueOf(entity).Elem()
	for _, f := range r.fields {
		if f.name == name {
			return v.Field(f.idx).Interface()
		}
	}
	return nil
}

func (r *BaseRepo[E, ID]) entityID(entity *E) interface{} {
	return r.fieldVal(entity, "id")
}

func (r *BaseRepo[E, ID]) entityScope(entity *E) interface{} {
	if r.scope == nil {
		return nil
	}
	return r.fieldVal(entity, r.scope.Column)
}

func (r *BaseRepo[E, ID]) ctxScope(ctx context.Context) interface{} {
	if r.scope == nil || r.scope.Source == nil {
		return nil
	}
	return r.scope.Source(ctx)
}

func (r *BaseRepo[E, ID]) colVals(entity *E, cols []string) []interface{} {
	vals := make([]interface{}, 0, len(cols))
	for _, c := range cols {
		if r.scope != nil && c == r.scope.Column {
			continue
		}
		vals = append(vals, r.fieldVal(entity, c))
	}
	return vals
}

// createActorArg prefers the request-scoped actor (WithActor) for created_by
// and falls back to the value the caller set on the entity, so entities that
// populate created_by themselves still satisfy NOT NULL audit columns when no
// actor is attached to the context.
func (r *BaseRepo[E, ID]) createActorArg(ctx context.Context, entity *E) interface{} {
	if a := Actor(ctx); a != "" {
		return a
	}
	return r.fieldVal(entity, "created_by")
}

func (r *BaseRepo[E, ID]) rebind(query string) string {
	return r.db.Rebind(query)
}

func (r *BaseRepo[E, ID]) exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return sqlext.GetQuerier(ctx, r.db).ExecContext(ctx, r.rebind(query), args...)
}

func (r *BaseRepo[E, ID]) get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlext.GetQuerier(ctx, r.db).GetContext(ctx, dest, r.rebind(query), args...)
}

func (r *BaseRepo[E, ID]) select_(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlext.GetQuerier(ctx, r.db).SelectContext(ctx, dest, r.rebind(query), args...)
}

func (r *BaseRepo[E, ID]) Create(ctx context.Context, entity *E) error {
	if entity == nil {
		return fmt.Errorf("entity is nil")
	}
	r.init()
	r.initType()

	if r.table.ValidateCreate != nil {
		if err := r.table.ValidateCreate(ctx, entity); err != nil {
			return err
		}
	}

	args := []interface{}{r.entityID(entity)}
	if r.scope != nil {
		args = append(args, r.entityScope(entity))
	}
	args = append(args, r.colVals(entity, r.q.createCols)...)
	if r.hasActor {
		args = append(args, r.createActorArg(ctx, entity))
	}

	if _, err := r.exec(ctx, r.q.create, args...); err != nil {
		return err
	}

	// Re-fetch so DB-computed columns (defaults, NOW() timestamps, generated id)
	// are reflected back into the returned entity.
	fargs := []interface{}{r.entityID(entity)}
	if r.scope != nil {
		fargs = append(fargs, r.entityScope(entity))
	}
	var fetched E
	if err := r.get(ctx, &fetched, r.q.findByID, fargs...); err == nil {
		*entity = fetched
	}
	return nil
}

func (r *BaseRepo[E, ID]) FindByID(ctx context.Context, id ID) (*E, error) {
	r.init()

	args := []interface{}{id}
	if r.scope != nil {
		args = append(args, r.ctxScope(ctx))
	}
	q := r.q.findByID
	if r.table.ExtraScope != nil {
		if frag, eargs := r.table.ExtraScope(ctx, len(args)+1); frag != "" {
			q += " AND " + frag
			args = append(args, eargs...)
		}
	}
	var entity E
	if err := r.get(ctx, &entity, q, args...); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *BaseRepo[E, ID]) FindByIDForUpdate(ctx context.Context, id ID) (*E, error) {
	r.init()

	args := []interface{}{id}
	if r.scope != nil {
		args = append(args, r.ctxScope(ctx))
	}
	q := r.q.findByID
	if r.table.ExtraScope != nil {
		if frag, eargs := r.table.ExtraScope(ctx, len(args)+1); frag != "" {
			q += " AND " + frag
			args = append(args, eargs...)
		}
	}
	var entity E
	if err := r.get(ctx, &entity, q+" FOR UPDATE", args...); err != nil {
		return nil, err
	}
	return &entity, nil
}

var placeholderRe = regexp.MustCompile(`\$(\d+)`)

// shiftPlaceholders renumbers $N parameters in clause by delta so a caller's
// WHERE fragment can be prefixed with additional bound arguments.
func shiftPlaceholders(clause string, delta int) string {
	return placeholderRe.ReplaceAllStringFunc(clause, func(m string) string {
		n, _ := strconv.Atoi(m[1:])
		return fmt.Sprintf("$%d", n+delta)
	})
}

// maxPlaceholder returns the highest $N parameter referenced in clause,
// or 0 when the clause binds no parameters.
func maxPlaceholder(clause string) int {
	max := 0
	for _, m := range placeholderRe.FindAllStringSubmatch(clause, -1) {
		n, _ := strconv.Atoi(m[1])
		if n > max {
			max = n
		}
	}
	return max
}

// normalizedLimit clamps a caller-supplied limit to the pagination bounds.
func normalizedLimit(limit int) int {
	if limit < 1 || limit > pagination.MaxLimit {
		return pagination.DefaultLimit
	}
	return limit
}

func (r *BaseRepo[E, ID]) Count(ctx context.Context, where string, args ...interface{}) (int, error) {
	r.init()

	if where != "" {
		if err := validateWhereClause(where); err != nil {
			return 0, err
		}
	}

	whereArgs := args
	if where != "" {
		where = "(" + where + ")"
	}
	if r.scope != nil {
		scopeWhere := r.scope.Column + "=$1"
		if where != "" {
			where = scopeWhere + " AND " + shiftPlaceholders(where, 1)
		} else {
			where = scopeWhere
		}
		whereArgs = append([]interface{}{r.ctxScope(ctx)}, args...)
	}
	if r.table.ExtraScope != nil {
		if frag, eargs := r.table.ExtraScope(ctx, maxPlaceholder(where)+1); frag != "" {
			if where == "" {
				where = frag
			} else {
				where = where + " AND " + frag
			}
			whereArgs = append(whereArgs, eargs...)
		}
	}
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.table.Name)
	if where != "" {
		query += " WHERE " + where
	}
	var total int
	if err := r.get(ctx, &total, query, whereArgs...); err != nil {
		return 0, err
	}
	return total, nil
}

func validateWhereClause(where string) error {
	if strings.ContainsAny(where, "'\";\x00") || strings.Contains(where, "--") || strings.Contains(where, "/*") {
		return fmt.Errorf("where clause contains unsafe characters")
	}
	return nil
}

func (r *BaseRepo[E, ID]) UpdateField(ctx context.Context, id ID, field string, value interface{}) error {
	r.init()
	r.initType()
	if !r.hasField(field) {
		return fmt.Errorf("unknown field: %s", field)
	}

	query := fmt.Sprintf("UPDATE %s SET %s=$1, updated_at=NOW() WHERE id=$2", r.table.Name, field)
	args := []interface{}{value, id}
	if r.hasUpdBy {
		query = fmt.Sprintf("UPDATE %s SET %s=$1, updated_by=$2, updated_at=NOW() WHERE id=$3", r.table.Name, field)
		args = []interface{}{value, actorArg(ctx), id}
	}
	if r.scope != nil {
		if r.hasUpdBy {
			query += " AND " + r.scope.Column + "=$4"
		} else {
			query += " AND " + r.scope.Column + "=$3"
		}
		args = append(args, r.ctxScope(ctx))
	}
	if r.table.ExtraScope != nil {
		if frag, eargs := r.table.ExtraScope(ctx, len(args)+1); frag != "" {
			query += " AND " + frag
			args = append(args, eargs...)
		}
	}
	res, err := r.exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// overlayExisting fills zero-valued UpdColumns of entity with the persisted
// row's values so a partial update body doesn't blank unchanged fields (e.g. a
// PUT carrying only {status} would otherwise reset santri_id to a zero uuid and
// trip the FK / NOT NULL constraints). Fields the caller explicitly sets to
// their zero value are intentionally left untouched, matching the conventional
// "ignore zero values on update" behaviour. The fetch is scoped (tenant +
// caller ExtraScope), so an out-of-scope update resolves to sql.ErrNoRows.
func (r *BaseRepo[E, ID]) overlayExisting(ctx context.Context, entity *E) error {
	id := r.entityID(entity)
	args := []interface{}{id}
	if r.scope != nil {
		args = append(args, r.ctxScope(ctx))
	}
	q := r.q.findByID
	if r.table.ExtraScope != nil {
		if frag, eargs := r.table.ExtraScope(ctx, len(args)+1); frag != "" {
			q += " AND " + frag
			args = append(args, eargs...)
		}
	}
	var existing E
	if err := r.get(ctx, &existing, q, args...); err != nil {
		return err
	}
	iv := reflect.ValueOf(entity).Elem()
	ev := reflect.ValueOf(&existing).Elem()
	for _, c := range r.table.UpdColumns {
		if r.scope != nil && c == r.scope.Column {
			continue
		}
		// Only restore zero-valued *relationship id* columns (those ending in
		// "_id"): a partial PUT that omits e.g. santri_id would otherwise reset
		// it to a zero uuid and trip FK / NOT NULL constraints. Scalar columns
		// (bools, ints, strings) are left as-is so a legitimate zero value the
		// caller explicitly sent (e.g. is_boarding=false) is preserved.
		if !strings.HasSuffix(c, "_id") {
			continue
		}
		idx := -1
		for _, f := range r.fields {
			if f.name == c {
				idx = f.idx
				break
			}
		}
		if idx < 0 {
			continue
		}
		inF := iv.Field(idx)
		if inF.IsZero() {
			inF.Set(ev.Field(idx))
		}
	}
	return nil
}

func (r *BaseRepo[E, ID]) Update(ctx context.Context, entity *E) error {
	if entity == nil {
		return fmt.Errorf("entity is nil")
	}
	r.init()
	r.initType()

	if err := r.overlayExisting(ctx, entity); err != nil {
		return err
	}

	if r.table.ValidateUpdate != nil {
		if err := r.table.ValidateUpdate(ctx, entity); err != nil {
			return err
		}
	}

	args := r.colVals(entity, r.q.updCols)
	q := r.q.update
	if r.hasUpdBy {
		if a := Actor(ctx); a != "" {
			args = append(args, a)
		} else {
			q = r.q.updateNoActor
		}
	}
	args = append(args, r.entityID(entity))
	if r.scope != nil {
		args = append(args, r.ctxScope(ctx))
	}
	if r.table.ExtraScope != nil {
		if frag, eargs := r.table.ExtraScope(ctx, len(args)+1); frag != "" {
			q += " AND " + frag
			args = append(args, eargs...)
		}
	}
	res, err := r.exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if r.table.ExtraScope != nil {
		if n, err := res.RowsAffected(); err == nil && n == 0 {
			return sql.ErrNoRows
		}
	}
	return nil
}

func (r *BaseRepo[E, ID]) Delete(ctx context.Context, id ID) error {
	r.init()

	args := []interface{}{id}
	if r.scope != nil {
		args = append(args, r.ctxScope(ctx))
	}
	q := r.q.delete
	if r.table.ExtraScope != nil {
		if frag, eargs := r.table.ExtraScope(ctx, len(args)+1); frag != "" {
			q += " AND " + frag
			args = append(args, eargs...)
		}
	}
	res, err := r.exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// allowedSortColumns returns columns safe to sort by: explicit SortCols plus
// real table columns and common meta columns.
func (r *BaseRepo[E, ID]) allowedSortColumns() map[string]string {
	allowed := make(map[string]string)
	for k, v := range r.table.SortCols {
		allowed[k] = v
	}
	for _, c := range r.table.Columns {
		allowed[c] = c
	}
	for _, c := range r.table.UpdColumns {
		allowed[c] = c
	}
	for _, c := range []string{"id", "created_at", "updated_at"} {
		allowed[c] = c
	}
	return allowed
}

// buildWhere composes the WHERE clause (scope + filters + search) and args.
func (r *BaseRepo[E, ID]) buildWhere(ctx context.Context, p pagination.Pagination) (string, []interface{}) {
	var conds []string
	var args []interface{}

	if r.scope != nil {
		args = append(args, r.ctxScope(ctx))
		conds = append(conds, fmt.Sprintf("%s=$%d", r.scope.Column, len(args)))
	}

	for _, fc := range r.table.FilterCols {
		val, ok := p.Filters[fc.Key]
		if !ok || val == "" {
			continue
		}
		if fc.Type == "date_range" {
			parts := strings.SplitN(val, ",", 2)
			if len(parts) == 2 {
				args = append(args, parts[0], parts[1])
				// Upper bound is exclusive < to+1day so an inclusive day range
				// (`to` treated as a calendar day) also covers same-day records on
				// timestamptz columns, not just `date` columns.
				conds = append(conds, fmt.Sprintf("%s >= $%d AND %s < ($%d::date + 1)", fc.Col, len(args)-1, fc.Col, len(args)))
			}
			continue
		}
		args = append(args, val)
		conds = append(conds, fmt.Sprintf("%s = $%d", fc.Col, len(args)))
	}

	if p.Search != "" && len(r.table.SearchCols) > 0 {
		var ors []string
		for _, c := range r.table.SearchCols {
			args = append(args, "%"+p.Search+"%")
			ors = append(ors, fmt.Sprintf("%s ILIKE $%d", c, len(args)))
		}
		conds = append(conds, "("+strings.Join(ors, " OR ")+")")
	}

	if r.table.ExtraScope != nil {
		frag, eargs := r.table.ExtraScope(ctx, len(args)+1)
		if frag != "" {
			args = append(args, eargs...)
			conds = append(conds, frag)
		}
	}

	if len(conds) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}

func (r *BaseRepo[E, ID]) List(ctx context.Context, p pagination.Pagination) ([]E, int, error) {
	r.init()

	where, args := r.buildWhere(ctx, p)

	var total int
	countQ := "SELECT COUNT(*) FROM " + r.table.Name + where
	if err := r.get(ctx, &total, countQ, args...); err != nil {
		return nil, 0, err
	}

	order := r.table.OrderBy
	if sc := r.allowedSortColumns(); p.Sort != "" {
		if col, ok := sc[p.Sort]; ok {
			dir := "ASC"
			if strings.EqualFold(p.Dir, "desc") {
				dir = "DESC"
			}
			order = fmt.Sprintf("%s %s", col, dir)
		}
	}

	limit := normalizedLimit(p.Limit)
	dataArgs := append(append([]interface{}{}, args...), limit, p.Offset())
	dataQ := fmt.Sprintf("SELECT * FROM %s%s ORDER BY %s LIMIT $%d OFFSET $%d",
		r.table.Name, where, order, len(args)+1, len(args)+2)
	var entities []E
	if err := r.select_(ctx, &entities, dataQ, dataArgs...); err != nil {
		return nil, 0, err
	}
	if entities == nil {
		entities = []E{}
	}
	return entities, total, nil
}

// JoinClause is a raw SQL JOIN clause used with ListDetail.
// Use table alias "e" for the main entity table.
type JoinClause struct {
	SQL string
}

// ExtraCol is an extra SELECT expression added when using ListDetail.
type ExtraCol struct {
	Expr  string // SQL expression, e.g. "COALESCE(v.c, 0)"
	Alias string // column alias, e.g. "venue_count"
}

// ListDetail performs a paginated list with extra columns from JOINs.
// The destination must be a pointer to a slice of structs that can accept
// the base entity columns (via e.*) plus the extra columns.
// Scope is applied automatically if configured on the BaseRepo.
//
// Example:
//
//	var data []domain.TenantWithCounts
//	total, err := repo.ListDetail(ctx, page, []crud.JoinClause{
//	    {SQL: "LEFT JOIN (SELECT tenant_id, COUNT(*) AS c FROM venues GROUP BY tenant_id) v ON v.tenant_id = e.id"},
//	}, []crud.ExtraCol{
//	    {Expr: "COALESCE(v.c, 0)", Alias: "venue_count"},
//	}, &data)
func (r *BaseRepo[E, ID]) ListDetail(ctx context.Context, page pagination.Pagination, joins []JoinClause, extras []ExtraCol, dest interface{}) (int, error) {
	r.init()

	joinParts := make([]string, len(joins))
	for i, j := range joins {
		joinParts[i] = j.SQL
	}
	extraParts := make([]string, len(extras))
	for i, e := range extras {
		extraParts[i] = fmt.Sprintf("%s AS %s", e.Expr, e.Alias)
	}

	joinsClause := strings.Join(joinParts, " ")
	selectExtras := strings.Join(extraParts, ", ")

	var whereClause string
	var args []interface{}
	paramIdx := 0
	if r.scope != nil {
		paramIdx++
		whereClause = fmt.Sprintf(" WHERE e.%s = $%d", r.scope.Column, paramIdx)
		args = append(args, r.ctxScope(ctx))
	}

	if r.table.ExtraScope != nil {
		frag, eargs := r.table.ExtraScope(ctx, paramIdx+1)
		if frag != "" {
			args = append(args, eargs...)
			if whereClause == "" {
				whereClause = " WHERE " + frag
			} else {
				whereClause += " AND " + frag
			}
			paramIdx += len(eargs)
		}
	}

	var total int
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s e %s%s", r.table.Name, joinsClause, whereClause)
	if err := r.get(ctx, &total, countSQL, args...); err != nil {
		return 0, err
	}

	paramIdx++
	limitIdx := paramIdx
	paramIdx++
	offsetIdx := paramIdx
	selectCols := "e.*"
	if selectExtras != "" {
		selectCols = "e.*, " + selectExtras
	}
	dataSQL := fmt.Sprintf("SELECT %s FROM %s e %s%s ORDER BY %s LIMIT $%d OFFSET $%d",
		selectCols, r.table.Name, joinsClause, whereClause, r.table.OrderBy, limitIdx, offsetIdx)
	dataArgs := append(args, normalizedLimit(page.Limit), page.Offset())

	if err := r.select_(ctx, dest, dataSQL, dataArgs...); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *BaseRepo[E, ID]) DB() *sqlx.DB { return r.db }

func (r *BaseRepo[E, ID]) TableInfo() TableInfo { return r.table }

func (r *BaseRepo[E, ID]) Scope() *ScopeConfig { return r.scope }
