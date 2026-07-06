package handler

import (
	"errors"
	"testing"
)

func TestDomainErrors(t *testing.T) {
	tests := []struct {
		err     error
		message string
	}{
		{ErrNotFound, "not found"},
		{ErrAlreadyExists, "already exists"},
		{ErrInvalidInput, "invalid input"},
		{ErrUnauthorized, "unauthorized"},
		{ErrForbidden, "forbidden"},
		{ErrPermissionDenied, "permission denied"},
		{ErrInsufficientBalance, "insufficient balance"},
		{ErrSlotNotAvailable, "slot not available"},
		{ErrSessionFull, "session is full"},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			if tt.err.Error() != tt.message {
				t.Errorf("Error() = %q, want %q", tt.err.Error(), tt.message)
			}
			if !errors.Is(tt.err, tt.err) {
				t.Errorf("errors.Is(%v, %v) = false", tt.err, tt.err)
			}
		})
	}
}

func TestErrComparison(t *testing.T) {
	err := ErrNotFound

	if !errors.Is(err, ErrNotFound) {
		t.Error("errors.Is(err, ErrNotFound) should be true")
	}
	if errors.Is(err, ErrAlreadyExists) {
		t.Error("errors.Is(err, ErrAlreadyExists) should be false")
	}
}
