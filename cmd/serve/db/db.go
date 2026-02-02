package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var (
	connString = flag.String("db-connection-string", "postgres://postgres:postgres@localhost/postgres", "Connection string for database")

	db *sqlx.DB
)

func Init() error {
	var err error

	db, err = sqlx.Connect("postgres", *connString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, *connString)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// FindContentByPath returns the content type for the given path.
// It handles cases where the path may or may not have a trailing slash.
// If path is "/foo", it will also check for "/foo/" in the database.
// For prefix matches (goimports), it will match subpaths like "/foo/bar".
// Returns "", nil if no matching path is found.
func FindContentByPath(ctx context.Context, path string) (string, error) {
	var contentType string
	err := db.GetContext(ctx, &contentType, `
		SELECT content_type FROM paths
		WHERE path = $1 OR path = $2 OR (prefix_match AND $1 LIKE path || '%')
		ORDER BY
			prefix_match ASC,
			LENGTH(path) DESC
		LIMIT 1
	`, path, path+"/")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return contentType, nil
}
