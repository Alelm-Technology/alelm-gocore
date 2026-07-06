package sqlext

import "strings"

func IsDuplicateKeyErr(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "UNIQUE constraint") || strings.Contains(msg, "Duplicate entry")
}
