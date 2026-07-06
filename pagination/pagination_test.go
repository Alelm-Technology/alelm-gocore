package pagination

import (
	"testing"
)

func TestDefaultLimit(t *testing.T) {
	if DefaultLimit != 20 {
		t.Errorf("DefaultLimit = %d, want 20", DefaultLimit)
	}
}

func TestMaxLimit(t *testing.T) {
	if MaxLimit != 100 {
		t.Errorf("MaxLimit = %d, want 100", MaxLimit)
	}
}

func TestOffset(t *testing.T) {
	tests := []struct {
		name string
		page Pagination
		want int
	}{
		{"page 1", Pagination{Page: 1, Limit: 20}, 0},
		{"page 2", Pagination{Page: 2, Limit: 20}, 20},
		{"page 3 limit 10", Pagination{Page: 3, Limit: 10}, 20},
		{"page 0 fallback", Pagination{Page: 0, Limit: 20}, 0},
		{"limit 0 fallback", Pagination{Page: 2, Limit: 0}, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.page.Offset(); got != tt.want {
				t.Errorf("Offset() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNewResponse(t *testing.T) {
	data := []string{"a", "b"}
	total := 45
	page := Pagination{Page: 2, Limit: 10}

	resp := NewResponse(data, total, page)

	if resp.Data.([]string)[0] != "a" {
		t.Errorf("Data[0] = %v, want 'a'", resp.Data)
	}
	if resp.Total != 45 {
		t.Errorf("Total = %d, want 45", resp.Total)
	}
	if resp.Page != 2 {
		t.Errorf("Page = %d, want 2", resp.Page)
	}
	if resp.Limit != 10 {
		t.Errorf("Limit = %d, want 10", resp.Limit)
	}
	if resp.TotalPages != 5 {
		t.Errorf("TotalPages = %d, want 5", resp.TotalPages)
	}
}

func TestNewResponseWithZeroLimit(t *testing.T) {
	data := []int{1, 2, 3}
	total := 30
	page := Pagination{Page: 1, Limit: 0}

	resp := NewResponse(data, total, page)

	if resp.Limit != DefaultLimit {
		t.Errorf("Limit = %d, want DefaultLimit(%d)", resp.Limit, DefaultLimit)
	}
}
