package util

import "time"

const (
	DateOnly     = "2006-01-02"
	DateTime     = "2006-01-02 15:04:05"
	DateTimeISO  = time.RFC3339
	DateID       = "02-01-2006"
	DateTimeID   = "02-01-2006 15:04"
)

func FormatDate(t time.Time) string {
	return t.Format(DateOnly)
}

func FormatDateTime(t time.Time) string {
	return t.Format(DateTime)
}

func FormatDateTimeISO(t time.Time) string {
	return t.Format(DateTimeISO)
}

func FormatDateID(t time.Time) string {
	return t.Format(DateID)
}

func FormatDateTimeID(t time.Time) string {
	return t.Format(DateTimeID)
}

func ParseDate(s string) (time.Time, error) {
	return time.Parse(DateOnly, s)
}

func ParseDateTime(s string) (time.Time, error) {
	return time.Parse(DateTime, s)
}

func ParseDateTimeISO(s string) (time.Time, error) {
	return time.Parse(DateTimeISO, s)
}

func TimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return pluralize(m, "minute") + " ago"
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return pluralize(h, "hour") + " ago"
	case d < 30*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return pluralize(days, "day") + " ago"
	default:
		return FormatDate(t)
	}
}

func pluralize(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
