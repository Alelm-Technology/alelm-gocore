package util

import (
	"testing"
)

func TestStrPtr(t *testing.T) {
	if got := StrPtr("hello"); *got != "hello" {
		t.Errorf("StrPtr('hello') = %v, want 'hello'", *got)
	}
	if got := StrPtr(""); got != nil {
		t.Error("StrPtr('') should return nil")
	}
}

func TestFloat64Ptr(t *testing.T) {
	if got := Float64Ptr(3.14); *got != 3.14 {
		t.Errorf("Float64Ptr(3.14) = %v, want 3.14", *got)
	}
	if got := Float64Ptr(0); got != nil {
		t.Error("Float64Ptr(0) should return nil")
	}
}

func TestIntPtr(t *testing.T) {
	if got := IntPtr(42); *got != 42 {
		t.Errorf("IntPtr(42) = %v, want 42", *got)
	}
	if got := IntPtr(0); got != nil {
		t.Error("IntPtr(0) should return nil")
	}
}

func TestBoolPtr(t *testing.T) {
	if got := BoolPtr(true); *got != true {
		t.Errorf("BoolPtr(true) = %v, want true", *got)
	}
	if got := BoolPtr(false); *got != false {
		t.Errorf("BoolPtr(false) = %v, want false", *got)
	}
}
