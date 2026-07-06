package util

import (
	"strings"
	"testing"
)

func TestNewUUID(t *testing.T) {
	id := NewUUID()
	if id == "" {
		t.Error("NewUUID() returned empty string")
	}
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		t.Errorf("NewUUID() = %q, expected 5 parts (dashed format)", id)
	}
}

func TestRandomString(t *testing.T) {
	s := RandomString(16)
	if len(s) != 16 {
		t.Errorf("RandomString(16) length = %d, want 16", len(s))
	}

	s1 := RandomString(8)
	s2 := RandomString(8)
	if s1 == s2 {
		t.Log("warning: RandomString generated same value twice (unlikely but possible)")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"shorter than max", "hello", 10, "hello"},
		{"equal to max", "hello", 5, "hello"},
		{"longer than max", "hello world", 5, "hello..."},
		{"empty string", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.input, tt.maxLen); got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if !Contains(slice, "a") {
		t.Error("Contains(slice, 'a') = false, want true")
	}
	if Contains(slice, "d") {
		t.Error("Contains(slice, 'd') = true, want false")
	}
	if Contains(nil, "a") {
		t.Error("Contains(nil, 'a') = true, want false")
	}
}
