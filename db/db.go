package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"flag"
	"fmt"
	"time"

	"chameth.com/chameth.com/features/metrics"
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

func Get[T any](ctx context.Context, query string, args ...any) (T, error) {
	metrics.LogQuery(ctx)
	var result T
	err := db.GetContext(ctx, &result, query, args...)
	return result, err
}

func Select[T any](ctx context.Context, query string, args ...any) ([]T, error) {
	metrics.LogQuery(ctx)
	var results []T
	err := db.SelectContext(ctx, &results, query, args...)
	return results, err
}

func Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	metrics.LogQuery(ctx)
	return db.ExecContext(ctx, query, args...)
}

func NamedExec(ctx context.Context, query string, arg any) (sql.Result, error) {
	metrics.LogQuery(ctx)
	return db.NamedExecContext(ctx, query, arg)
}

func QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	metrics.LogQuery(ctx)
	return db.QueryRowContext(ctx, query, args...)
}

func Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	metrics.LogQuery(ctx)
	return db.QueryContext(ctx, query, args...)
}

func FindContentByPath(ctx context.Context, path string) (string, error) {
	contentType, err := Get[string](ctx, `
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
