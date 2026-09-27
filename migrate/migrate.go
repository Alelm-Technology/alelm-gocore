package migrate

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func Run(db *sql.DB, sourcePath string) error {
	goose.SetDialect("postgres")
	if err := goose.Up(db, sourcePath); err != nil && err != goose.ErrNoNextVersion {
		return fmt.Errorf("failed to run migration: %w", err)
	}
	return nil
}
