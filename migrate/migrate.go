package migrate

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Run(db *sql.DB, sourcePath string) {
	driver, err := migratepg.WithInstance(db, &migratepg.Config{})
	if err != nil {
		slog.Error("failed to create migration driver", "error", err)
		os.Exit(1)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+sourcePath, "postgres", driver)
	if err != nil {
		slog.Error("failed to init migration", "error", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("failed to run migration", "error", err)
		os.Exit(1)
	}

	slog.Info("migrations applied")
}
