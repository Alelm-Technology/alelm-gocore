package util

import (
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	now := time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC)
	if got := FormatDate(now); got != "2026-07-06" {
		t.Errorf("FormatDate() = %q, want '2026-07-06'", got)
	}
}

func TestFormatDateTime(t *testing.T) {
	now := time.Date(2026, 7, 6, 10, 5, 30, 0, time.UTC)
	if got := FormatDateTime(now); got != "2026-07-06 10:05:30" {
		t.Errorf("FormatDateTime() = %q, want '2026-07-06 10:05:30'", got)
	}
}

func TestParseDate(t *testing.T) {
	got, err := ParseDate("2026-07-06")
	if err != nil {
		t.Fatal(err)
	}
	if got.Year() != 2026 || got.Month() != 7 || got.Day() != 6 {
		t.Errorf("ParseDate() = %v", got)
	}
}

func TestParseDateInvalid(t *testing.T) {
	_, err := ParseDate("not-a-date")
	if err == nil {
		t.Error("ParseDate('not-a-date') should error")
	}
}

func TestTimeAgoJustNow(t *testing.T) {
	if got := TimeAgo(time.Now()); got != "just now" {
		t.Errorf("TimeAgo(now) = %q, want 'just now'", got)
	}
}

func TestFormatDateID(t *testing.T) {
	now := time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC)
	if got := FormatDateID(now); got != "06-07-2026" {
		t.Errorf("FormatDateID() = %q, want '06-07-2026'", got)
	}
}
