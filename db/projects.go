package db

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/metrics"
)

// GetAllProjectSections returns all project sections ordered by sort.
func GetAllProjectSections(ctx context.Context) ([]ProjectSection, error) {
	metrics.LogQuery(ctx)
	var sections []ProjectSection
	err := db.SelectContext(ctx, &sections, "SELECT id, name, sort, description FROM project_sections ORDER BY sort")
	if err != nil {
		return nil, err
	}
	return sections, nil
}

// GetProjectsInSection returns all projects in a section ordered by pinned (descending), then name (case-insensitive).
func GetProjectsInSection(ctx context.Context, sectionID int) ([]Project, error) {
	metrics.LogQuery(ctx)
	var projects []Project
	err := db.SelectContext(ctx, &projects, "SELECT id, section, name, icon, pinned, description FROM projects WHERE section = $1 AND published = true ORDER BY pinned DESC, LOWER(name)", sectionID)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// GetAllProjects returns all published projects.
func GetAllProjects(ctx context.Context) ([]Project, error) {
	metrics.LogQuery(ctx)
	var projects []Project
	err := db.SelectContext(ctx, &projects, "SELECT id, section, name, icon, pinned, description FROM projects WHERE published = true ORDER BY section, pinned DESC, LOWER(name)")
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// GetDraftProjects returns all unpublished projects.
func GetDraftProjects(ctx context.Context) ([]Project, error) {
	metrics.LogQuery(ctx)
	var projects []Project
	err := db.SelectContext(ctx, &projects, "SELECT id, section, name, icon, pinned, description FROM projects WHERE published = false ORDER BY section, pinned DESC, LOWER(name)")
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProjectByID returns a project for the given ID.
func GetProjectByID(ctx context.Context, id int) (*Project, error) {
	metrics.LogQuery(ctx)
	var project Project
	err := db.GetContext(ctx, &project, "SELECT id, section, name, icon, pinned, description, published FROM projects WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// CreateProject creates a new project in the database and returns its ID.
// The project is created with the first section by sort order, unpublished, not pinned.
func CreateProject(ctx context.Context, name string) (int, error) {
	metrics.LogQuery(ctx)
	sections, err := GetAllProjectSections(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get sections: %w", err)
	}
	if len(sections) == 0 {
		return 0, fmt.Errorf("no sections available")
	}
	defaultSection := sections[0].ID

	var id int
	err = db.QueryRowContext(ctx, `
		INSERT INTO projects (section, name, icon, pinned, description, published)
		VALUES ($1, $2, '', false, '', false)
		RETURNING id
	`, defaultSection, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create project: %w", err)
	}
	return id, nil
}

// UpdateProject updates a project in the database.
func UpdateProject(ctx context.Context, id int, name, icon, description string, section int, pinned, published bool) error {
	metrics.LogQuery(ctx)
	_, err := db.ExecContext(ctx, `
		UPDATE projects
		SET section = $1, name = $2, icon = $3, pinned = $4, description = $5, published = $6
		WHERE id = $7
	`, section, name, icon, pinned, description, published, id)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}
