package crud

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"

	"github.com/alelmtech/gocore/pagination"
)

type itemRow struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

func newMockRepo(t *testing.T, table TableInfo, scope *ScopeConfig) (*BaseRepo[itemRow, string], sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	r := NewBaseRepo[itemRow, string](sqlx.NewDb(db, "sqlmock"), table, scope)
	r.init()
	r.initType()
	return r, mock
}

func itemTable(extra func(ctx context.Context, nextParam int) (string, []interface{})) TableInfo {
	return TableInfo{
		Name:       "items",
		Columns:    []string{"name"},
		UpdColumns: []string{"name"},
		OrderBy:    "name",
		ExtraScope: extra,
	}
}

func TestShiftPlaceholders(t *testing.T) {
	cases := []struct {
		clause string
		delta  int
		want   string
	}{
		{"status = $1", 1, "status = $2"},
		{"a = $1 AND b = $2", 2, "a = $3 AND b = $4"},
		{"c = $10", 1, "c = $11"},
		{"no params", 3, "no params"},
	}
	for _, c := range cases {
		if got := shiftPlaceholders(c.clause, c.delta); got != c.want {
			t.Errorf("shiftPlaceholders(%q, %d) = %q, want %q", c.clause, c.delta, got, c.want)
		}
	}
}

func TestMaxPlaceholder(t *testing.T) {
	cases := []struct {
		clause string
		want   int
	}{
		{"status = $1", 1},
		{"a = $1 OR b = $7 OR c = $3", 7},
		{"", 0},
		{"no params", 0},
		{"c = $10 AND d = $2", 10},
	}
	for _, c := range cases {
		if got := maxPlaceholder(c.clause); got != c.want {
			t.Errorf("maxPlaceholder(%q) = %d, want %d", c.clause, got, c.want)
		}
	}
}

func TestNormalizedLimit(t *testing.T) {
	cases := []struct {
		limit int
		want  int
	}{
		{-1, 20},
		{0, 20},
		{1, 1},
		{100, 100},
		{101, 20},
		{1000000, 20},
	}
	for _, c := range cases {
		if got := normalizedLimit(c.limit); got != c.want {
			t.Errorf("normalizedLimit(%d) = %d, want %d", c.limit, got, c.want)
		}
	}
}

func TestCountAppliesExtraScope(t *testing.T) {
	r, mock := newMockRepo(t, itemTable(func(_ context.Context, next int) (string, []interface{}) {
		return "viewer_id=$" + strconv.Itoa(next), []interface{}{"u9"}
	}), nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM items WHERE (status = $1) AND viewer_id=$2")).
		WithArgs("published", "u9").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	got, err := r.Count(context.Background(), "status = $1", "published")
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if got != 3 {
		t.Errorf("Count = %d, want 3", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestCountScopeRenumbersWhere(t *testing.T) {
	r, mock := newMockRepo(t, itemTable(nil), &ScopeConfig{
		Column: "tenant_id",
		Source: func(context.Context) interface{} { return "t1" },
	})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM items WHERE tenant_id=$1 AND (status = $2)")).
		WithArgs("t1", "published").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	got, err := r.Count(context.Background(), "status = $1", "published")
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if got != 1 {
		t.Errorf("Count = %d, want 1", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestCountScopeParenthesizesCallerWhere(t *testing.T) {
	r, mock := newMockRepo(t, itemTable(nil), &ScopeConfig{
		Column: "tenant_id",
		Source: func(context.Context) interface{} { return "t1" },
	})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM items WHERE tenant_id=$1 AND (a = $2 OR b = $3)")).
		WithArgs("t1", "x", "y").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	if _, err := r.Count(context.Background(), "a = $1 OR b = $2", "x", "y"); err != nil {
		t.Fatalf("Count: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUpdateFieldAppliesExtraScopeAndErrNoRows(t *testing.T) {
	extraCalled := false
	r, mock := newMockRepo(t, itemTable(func(_ context.Context, next int) (string, []interface{}) {
		extraCalled = true
		if next != 3 {
			t.Errorf("ExtraScope nextParam = %d, want 3", next)
		}
		return "viewer_id=$3", []interface{}{"u9"}
	}), nil)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE items SET name=$1, updated_at=NOW() WHERE id=$2 AND viewer_id=$3")).
		WithArgs("renamed", "id1", "u9").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := r.UpdateField(context.Background(), "id1", "name", "renamed")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("UpdateField err = %v, want sql.ErrNoRows", err)
	}
	if !extraCalled {
		t.Error("ExtraScope was not invoked")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUpdateFieldScopedHitSucceeds(t *testing.T) {
	r, mock := newMockRepo(t, itemTable(nil), nil)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE items SET name=$1, updated_at=NOW() WHERE id=$2")).
		WithArgs("renamed", "id1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := r.UpdateField(context.Background(), "id1", "name", "renamed"); err != nil {
		t.Fatalf("UpdateField: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestDeleteMissingRowReturnsErrNoRows(t *testing.T) {
	r, mock := newMockRepo(t, itemTable(nil), nil)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM items WHERE id=$1")).
		WithArgs("missing").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := r.Delete(context.Background(), "missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Delete err = %v, want sql.ErrNoRows", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestDeleteAppliesExtraScope(t *testing.T) {
	r, mock := newMockRepo(t, itemTable(func(_ context.Context, next int) (string, []interface{}) {
		return "viewer_id=$" + strconv.Itoa(next), []interface{}{"u9"}
	}), nil)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM items WHERE id=$1 AND viewer_id=$2")).
		WithArgs("id1", "u9").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := r.Delete(context.Background(), "id1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestListNormalizesOversizedLimit(t *testing.T) {
	r, mock := newMockRepo(t, itemTable(nil), nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM items")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(500))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM items ORDER BY name LIMIT $1 OFFSET $2")).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "a"))

	page := pagination.Pagination{Page: 1, Limit: 1000000}
	entities, total, err := r.List(context.Background(), page)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 500 || len(entities) != 1 {
		t.Errorf("List total=%d len=%d, want 500/1", total, len(entities))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestListDetailWithoutExtrasBuildsValidSelect(t *testing.T) {
	r, mock := newMockRepo(t, itemTable(nil), nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM items e ")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT e.* FROM items e  ORDER BY name LIMIT $1 OFFSET $2")).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "a"))

	var data []itemRow
	page := pagination.Pagination{Page: 1, Limit: 5000}
	total, err := r.ListDetail(context.Background(), page, nil, nil, &data)
	if err != nil {
		t.Fatalf("ListDetail: %v", err)
	}
	if total != 1 || len(data) != 1 {
		t.Errorf("ListDetail total=%d len=%d, want 1/1", total, len(data))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
