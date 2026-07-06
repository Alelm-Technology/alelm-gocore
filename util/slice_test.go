package util

import (
	"testing"
)

func TestMap(t *testing.T) {
	input := []int{1, 2, 3}
	got := Map(input, func(i int) string { return string(rune('a' + i - 1)) })

	if len(got) != 3 {
		t.Fatalf("Map length = %d, want 3", len(got))
	}
	if got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("Map = %v, want ['a', 'b', 'c']", got)
	}
}

func TestFilter(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	got := Filter(input, func(i int) bool { return i%2 == 0 })

	if len(got) != 2 {
		t.Fatalf("Filter length = %d, want 2", len(got))
	}
	if got[0] != 2 || got[1] != 4 {
		t.Errorf("Filter = %v, want [2, 4]", got)
	}
}

func TestSliceContains(t *testing.T) {
	slice := []int{10, 20, 30}

	if !SliceContains(slice, 20) {
		t.Error("SliceContains(slice, 20) = false, want true")
	}
	if SliceContains(slice, 99) {
		t.Error("SliceContains(slice, 99) = true, want false")
	}
	if SliceContains(nil, 1) {
		t.Error("SliceContains(nil, 1) = true, want false")
	}
}

func TestUnique(t *testing.T) {
	input := []int{1, 2, 2, 3, 3, 3, 4}
	got := Unique(input)

	if len(got) != 4 {
		t.Fatalf("Unique length = %d, want 4", len(got))
	}
	expected := []int{1, 2, 3, 4}
	for i, v := range expected {
		if got[i] != v {
			t.Errorf("Unique[%d] = %d, want %d", i, got[i], v)
		}
	}
}

func TestUniqueEmpty(t *testing.T) {
	got := Unique([]int{})
	if len(got) != 0 {
		t.Errorf("Unique(empty) length = %d, want 0", len(got))
	}
}
