package sqlext

import (
	"errors"
	"testing"
)

func TestIsDuplicateKeyErr(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"postgres duplicate key", errors.New(`pq: duplicate key value violates unique constraint "users_email_key"`), true},
		{"postgres unique constraint", errors.New(`pq: duplicate key value violates unique constraint "users_phone_key"`), true},
		{"mysql duplicate entry", errors.New("Error 1062: Duplicate entry 'test@test.com' for key 'users.email'"), true},
		{"not a duplicate key error", errors.New("some other error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDuplicateKeyErr(tt.err); got != tt.want {
				t.Errorf("IsDuplicateKeyErr() = %v, want %v", got, tt.want)
			}
		})
	}
}
