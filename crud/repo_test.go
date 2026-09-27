package crud

import (
	"strings"
	"testing"
)

type auditRow struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	CreatedBy string `db:"created_by"`
	UpdatedBy string `db:"updated_by"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type scopedAuditRow struct {
	ID        string `db:"id"`
	TenantID  string `db:"tenant_id"`
	Name      string `db:"name"`
	CreatedBy string `db:"created_by"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func insertCols(query string) []string {
	head := query[strings.Index(query, "(")+1 : strings.Index(query, ") VALUES")]
	parts := strings.Split(head, ", ")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

func setClauses(query string) []string {
	head := query[strings.Index(query, "SET ")+4 : strings.Index(query, " WHERE ")]
	return strings.Split(head, ", ")
}

func assertNoDuplicate(t *testing.T, cols []string) {
	t.Helper()
	seen := map[string]bool{}
	for _, c := range cols {
		col := c
		if i := strings.Index(col, "="); i >= 0 {
			col = col[:i]
		}
		if seen[col] {
			t.Errorf("duplicate column %q in query columns %v", col, cols)
		}
		seen[col] = true
	}
}

func TestCreateNoDuplicateAuditColumns(t *testing.T) {
	r := NewBaseRepo[auditRow, string](nil, TableInfo{
		Name:       "teams",
		Columns:    []string{"name", "created_by", "created_at", "updated_at"},
		UpdColumns: []string{"name", "created_by", "updated_by", "created_at", "updated_at"},
		OrderBy:    "created_at DESC",
	}, nil)
	r.init()
	r.initType()

	cols := insertCols(r.q.create)
	assertNoDuplicate(t, cols)
	if !strings.Contains(r.q.create, "created_at, updated_at") {
		t.Errorf("expected NOW() timestamps in INSERT, got %s", r.q.create)
	}
	wantParams := 1 + len(r.q.createCols) + 1 // id + bound cols + actor
	if got := strings.Count(r.q.create, "$"); got != wantParams {
		t.Errorf("INSERT has %d placeholders, want %d: %s", got, wantParams, r.q.create)
	}
	if !r.hasActor {
		t.Error("expected hasActor for entity with created_by")
	}

	sets := setClauses(r.q.update)
	assertNoDuplicate(t, sets)
	if len(r.q.updCols) != 2 {
		t.Errorf("updCols = %v, want [name created_by]", r.q.updCols)
	}
}

func TestCreateScopedNoDuplicateColumns(t *testing.T) {
	r := NewBaseRepo[scopedAuditRow, string](nil, TableInfo{
		Name:       "venues",
		Columns:    []string{"tenant_id", "name", "created_by", "created_at", "updated_at"},
		UpdColumns: []string{"tenant_id", "name", "created_at", "updated_at"},
		OrderBy:    "created_at DESC",
	}, &ScopeConfig{Column: "tenant_id"})
	r.init()
	r.initType()

	cols := insertCols(r.q.create)
	assertNoDuplicate(t, cols)
	if cols[1] != "tenant_id" {
		t.Errorf("scope column must be bound once at index 1, got %v", cols)
	}
	if !strings.Contains(r.q.create, "VALUES ($1, $2, $3,") {
		t.Errorf("expected id=$1 scope=$2 first col=$3, got %s", r.q.create)
	}
	wantParams := 1 + 1 + len(r.q.createCols) + 1 // id + scope + bound cols + actor
	if got := strings.Count(r.q.create, "$"); got != wantParams {
		t.Errorf("INSERT has %d placeholders, want %d: %s", got, wantParams, r.q.create)
	}

	sets := setClauses(r.q.update)
	assertNoDuplicate(t, sets)
	if len(r.q.updCols) != 1 || r.q.updCols[0] != "name" {
		t.Errorf("updCols = %v, want [name]", r.q.updCols)
	}
}

func TestCreateWithoutAuditFields(t *testing.T) {
	type plain struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}
	r := NewBaseRepo[plain, string](nil, TableInfo{
		Name:       "items",
		Columns:    []string{"name"},
		UpdColumns: []string{"name"},
		OrderBy:    "name",
	}, nil)
	r.init()
	r.initType()

	cols := insertCols(r.q.create)
	assertNoDuplicate(t, cols)
	if len(cols) != 2 {
		t.Errorf("INSERT columns = %v, want [id name]", cols)
	}
	if strings.Count(r.q.create, "$") != 2 {
		t.Errorf("INSERT placeholders mismatch: %s", r.q.create)
	}
	if r.hasActor || r.hasTS {
		t.Error("plain entity must not be treated as having audit columns")
	}
}
