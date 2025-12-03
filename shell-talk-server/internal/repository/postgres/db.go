package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// NewDB creates a new database connection.
func NewDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// RunMigrations applies all pending database migrations.
func RunMigrations(dataSourceName string) error {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return fmt.Errorf("could not open db for migration: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migrate driver: %w", err)
	}

	// The path is relative to the execution directory.
	// When running from `cmd/server`, it should be `../../internal/repository/postgres/migrations`
	// A more robust solution would use absolute paths or embed.
	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/repository/postgres/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("could not apply migrations: %w", err)
	}

	return nil
}
