package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	t.Run("returns value when set", func(t *testing.T) {
		os.Setenv("TEST_KEY", "hello")
		defer os.Unsetenv("TEST_KEY")

		if got := GetEnv("TEST_KEY", "fallback"); got != "hello" {
			t.Errorf("GetEnv() = %q, want 'hello'", got)
		}
	})

	t.Run("returns fallback when not set", func(t *testing.T) {
		if got := GetEnv("NONEXISTENT_KEY_XYZ", "fallback"); got != "fallback" {
			t.Errorf("GetEnv() = %q, want 'fallback'", got)
		}
	})

	t.Run("returns empty when set to empty string", func(t *testing.T) {
		os.Setenv("TEST_EMPTY", "")
		defer os.Unsetenv("TEST_EMPTY")

		if got := GetEnv("TEST_EMPTY", "fallback"); got != "fallback" {
			t.Errorf("GetEnv() = %q, want 'fallback'", got)
		}
	})
}
