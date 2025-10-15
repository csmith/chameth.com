package main

import (
	"database/sql"
	"embed"
	"errors"
	"flag"
	"fmt"

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

func initDatabase() error {
	var err error

	db, err = sqlx.Connect("postgres", *connString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

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

// findContentBySlug returns the content type for the given slug.
// It handles cases where the slug may or may not have a trailing slash.
// If slug is "/foo", it will also check for "/foo/" in the database.
// Returns "", nil if no matching slug is found.
func findContentBySlug(slug string) (string, error) {
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

// getPoemBySlug returns a poem for the given slug.
// It handles cases where the slug may or may not have a trailing slash.
// Returns nil if no poem is found with that slug.
func getPoemBySlug(slug string) (*Poem, error) {
	var poem Poem
	err := db.Get(&poem, "SELECT slug, title, poem, notes, published, modified FROM poems WHERE slug = $1 OR slug = $2", slug, slug+"/")
	if err != nil {
		return nil, err
	}
	return &poem, nil
}

// getSnippetBySlug returns a snippet for the given slug.
// It handles cases where the slug may or may not have a trailing slash.
// Returns nil if no snippet is found with that slug.
func getSnippetBySlug(slug string) (*Snippet, error) {
	var snippet Snippet
	err := db.Get(&snippet, "SELECT slug, title, topic, content FROM snippets WHERE slug = $1 OR slug = $2", slug, slug+"/")
	if err != nil {
		return nil, err
	}
	return &snippet, nil
}

// getStaticPageBySlug returns a static page for the given slug.
// It handles cases where the slug may or may not have a trailing slash.
// Returns nil if no static page is found with that slug.
func getStaticPageBySlug(slug string) (*StaticPage, error) {
	var page StaticPage
	err := db.Get(&page, "SELECT id, slug, title, content FROM staticpages WHERE slug = $1 OR slug = $2", slug, slug+"/")
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// getAllSnippets returns all snippets without their content.
func getAllSnippets() ([]Snippet, error) {
	var snippets []Snippet
	err := db.Select(&snippets, "SELECT slug, title, topic FROM snippets ORDER BY topic, title")
	if err != nil {
		return nil, err
	}
	return snippets, nil
}

// getAllProjectSections returns all project sections ordered by sort.
func getAllProjectSections() ([]ProjectSection, error) {
	var sections []ProjectSection
	err := db.Select(&sections, "SELECT id, name, sort, description FROM project_sections ORDER BY sort")
	if err != nil {
		return nil, err
	}
	return sections, nil
}

// getProjectsInSection returns all projects in a section ordered by pinned (descending), then name (case-insensitive).
func getProjectsInSection(sectionID int) ([]Project, error) {
	var projects []Project
	err := db.Select(&projects, "SELECT id, section, name, icon, pinned, description FROM projects WHERE section = $1 ORDER BY pinned DESC, LOWER(name)", sectionID)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// getMediaBySlug returns media for the given slug.
// Returns nil if no media is found with that slug.
func getMediaBySlug(slug string) (*Media, error) {
	var media Media
	err := db.Get(&media, `
		SELECT m.id, m.content_type, m.original_filename, m.data
		FROM media m
		JOIN media_relations mr ON m.id = mr.media_id
		WHERE mr.slug = $1
	`, slug)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// getAllPrints returns all prints ordered by name.
func getAllPrints() ([]Print, error) {
	var prints []Print
	err := db.Select(&prints, "SELECT id, name, description FROM prints ORDER BY name")
	if err != nil {
		return nil, err
	}
	return prints, nil
}

// getPrintLinks returns all links for a given print ID.
func getPrintLinks(printID int) ([]PrintLink, error) {
	var links []PrintLink
	err := db.Select(&links, "SELECT id, print_id, name, address FROM prints_links WHERE print_id = $1", printID)
	if err != nil {
		return nil, err
	}
	return links, nil
}

// getMediaRelationsForEntity returns all media relations for a given entity type and ID.
func getMediaRelationsForEntity(entityType string, entityID int) ([]MediaRelationWithDetails, error) {
	var relations []MediaRelationWithDetails
	err := db.Select(&relations, `
		SELECT
			mr.slug, mr.media_id, mr.description, mr.caption, mr.role, mr.entity_type, mr.entity_id,
			m.id, m.content_type, m.original_filename, m.width, m.height
		FROM media_relations mr
		JOIN media m ON mr.media_id = m.id
		WHERE mr.entity_type = $1 AND mr.entity_id = $2
	`, entityType, entityID)
	if err != nil {
		return nil, err
	}
	return relations, nil
}
