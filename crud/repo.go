package crud

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
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
		create   string
		findByID string
		update   string
		delete   string
		listCnt  string
		listData string
	}
	qOnce sync.Once

	fields  []fieldInfo
	fOnce   sync.Once
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
		numUpd := len(r.table.UpdColumns)

		{
			cols := []string{"id"}
			params := []string{"$1"}
			if hasScope {
				cols = append(cols, r.scope.Column)
				params = append(params, "$2")
			}
			for _, c := range r.table.Columns {
				cols = append(cols, c)
				params = append(params, fmt.Sprintf("$%d", len(params)+1))
			}
			if hasTS {
				cols = append(cols, "created_at", "updated_at")
				params = append(params, "NOW()", "NOW()")
			}
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
			setParts := make([]string, 0, numUpd+1)
			for _, c := range r.table.UpdColumns {
				setParts = append(setParts, fmt.Sprintf("%s=$%d", c, paramIdx))
				paramIdx++
			}
			if hasTS {
				setParts = append(setParts, "updated_at=NOW()")
			}
			idParam := paramIdx
			paramIdx++
			if hasScope {
				r.q.update = fmt.Sprintf("UPDATE %s SET %s WHERE id=$%d AND %s=$%d",
					r.table.Name, strings.Join(setParts, ", "), idParam, r.scope.Column, paramIdx)
			} else {
				r.q.update = fmt.Sprintf("UPDATE %s SET %s WHERE id=$%d",
					r.table.Name, strings.Join(setParts, ", "), idParam)
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
	vals := make([]interface{}, len(cols))
	for i, c := range cols {
		vals[i] = r.fieldVal(entity, c)
	}
	return vals
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
	r.init()
	r.initType()

	args := []interface{}{r.entityID(entity)}
	if r.scope != nil {
		args = append(args, r.entityScope(entity))
	}
	args = append(args, r.colVals(entity, r.table.Columns)...)

	_, err := r.exec(ctx, r.q.create, args...)
	return err
}

func (r *BaseRepo[E, ID]) FindByID(ctx context.Context, id ID) (*E, error) {
	r.init()

	args := []interface{}{id}
	if r.scope != nil {
		args = append(args, r.ctxScope(ctx))
	}
	var entity E
	if err := r.get(ctx, &entity, r.q.findByID, args...); err != nil {
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
	var entity E
	if err := r.get(ctx, &entity, r.q.findByID+" FOR UPDATE", args...); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *BaseRepo[E, ID]) Count(ctx context.Context, where string, args ...interface{}) (int, error) {
	r.init()

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.table.Name)
	if r.scope != nil {
		scopeWhere := r.scope.Column + "=$1"
		if where != "" {
			where = scopeWhere + " AND " + where
		} else {
			where = scopeWhere
		}
		args = append([]interface{}{r.ctxScope(ctx)}, args...)
	}
	if where != "" {
		query += " WHERE " + where
	}
	var total int
	if err := r.get(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *BaseRepo[E, ID]) UpdateField(ctx context.Context, id ID, field string, value interface{}) error {
	r.init()

	query := fmt.Sprintf("UPDATE %s SET %s=$1, updated_at=NOW() WHERE id=$2", r.table.Name, field)
	args := []interface{}{value, id}
	if r.scope != nil {
		query += " AND " + r.scope.Column + "=$3"
		args = append(args, r.ctxScope(ctx))
	}
	_, err := r.exec(ctx, query, args...)
	return err
}

func (r *BaseRepo[E, ID]) Update(ctx context.Context, entity *E) error {
	r.init()
	r.initType()

	args := r.colVals(entity, r.table.UpdColumns)
	args = append(args, r.entityID(entity))
	if r.scope != nil {
		args = append(args, r.ctxScope(ctx))
	}
	_, err := r.exec(ctx, r.q.update, args...)
	return err
}

func (r *BaseRepo[E, ID]) Delete(ctx context.Context, id ID) error {
	r.init()

	args := []interface{}{id}
	if r.scope != nil {
		args = append(args, r.ctxScope(ctx))
	}
	_, err := r.exec(ctx, r.q.delete, args...)
	return err
}

func (r *BaseRepo[E, ID]) List(ctx context.Context, page pagination.Pagination) ([]E, int, error) {
	r.init()

	var total int
	args := []interface{}{}
	if r.scope != nil {
		args = append(args, r.ctxScope(ctx))
	}
	if err := r.get(ctx, &total, r.q.listCnt, args...); err != nil {
		return nil, 0, err
	}

	dataArgs := make([]interface{}, len(args))
	copy(dataArgs, args)
	dataArgs = append(dataArgs, page.Limit, page.Offset())

	var entities []E
	if err := r.select_(ctx, &entities, r.q.listData, dataArgs...); err != nil {
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

	var total int
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s e %s%s", r.table.Name, joinsClause, whereClause)
	if err := r.get(ctx, &total, countSQL, args...); err != nil {
		return 0, err
	}

	paramIdx++
	limitIdx := paramIdx
	paramIdx++
	offsetIdx := paramIdx
	dataSQL := fmt.Sprintf("SELECT e.*, %s FROM %s e %s%s ORDER BY %s LIMIT $%d OFFSET $%d",
		selectExtras, r.table.Name, joinsClause, whereClause, r.table.OrderBy, limitIdx, offsetIdx)
	dataArgs := append(args, page.Limit, page.Offset())

	if err := r.select_(ctx, dest, dataSQL, dataArgs...); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *BaseRepo[E, ID]) DB() *sqlx.DB { return r.db }

func (r *BaseRepo[E, ID]) TableInfo() TableInfo { return r.table }

func (r *BaseRepo[E, ID]) Scope() *ScopeConfig { return r.scope }
