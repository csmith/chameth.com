package db

import (
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

// FindContentBySlug returns the content type for the given slug.
// It handles cases where the slug may or may not have a trailing slash.
// If slug is "/foo", it will also check for "/foo/" in the database.
// Returns "", nil if no matching slug is found.
func FindContentBySlug(slug string) (string, error) {
	var contentType string
	err := db.Get(&contentType, "SELECT content_type FROM slugs WHERE slug = $1 OR slug = $2", slug, slug+"/")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return contentType, nil
}
