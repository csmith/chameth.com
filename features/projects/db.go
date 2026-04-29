package projects

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
)

func GetAllProjectSections(ctx context.Context) ([]ProjectSection, error) {
	return db.Select[ProjectSection](ctx, "SELECT id, name, sort, description FROM project_sections ORDER BY sort")
}

func GetProjectsInSection(ctx context.Context, sectionID int) ([]Project, error) {
	return db.Select[Project](ctx, "SELECT id, section, name, icon, pinned, description FROM projects WHERE section = $1 AND published = true ORDER BY pinned DESC, LOWER(name)", sectionID)
}

func GetAllProjects(ctx context.Context) ([]Project, error) {
	return db.Select[Project](ctx, "SELECT id, section, name, icon, pinned, description FROM projects WHERE published = true ORDER BY section, pinned DESC, LOWER(name)")
}

func GetDraftProjects(ctx context.Context) ([]Project, error) {
	return db.Select[Project](ctx, "SELECT id, section, name, icon, pinned, description FROM projects WHERE published = false ORDER BY section, pinned DESC, LOWER(name)")
}

func GetProjectByID(ctx context.Context, id int) (*Project, error) {
	project, err := db.Get[Project](ctx, "SELECT id, section, name, icon, pinned, description, published FROM projects WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func CreateProject(ctx context.Context, name string) (int, error) {
	sections, err := GetAllProjectSections(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get sections: %w", err)
	}
	if len(sections) == 0 {
		return 0, fmt.Errorf("no sections available")
	}
	defaultSection := sections[0].ID

	var id int
	err = db.QueryRow(ctx, `
		INSERT INTO projects (section, name, icon, pinned, description, published)
		VALUES ($1, $2, '', false, '', false)
		RETURNING id
	`, defaultSection, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create project: %w", err)
	}
	return id, nil
}

func UpdateProject(ctx context.Context, id int, name, icon, description string, section int, pinned, published bool) error {
	_, err := db.Exec(ctx, `
		UPDATE projects
		SET section = $1, name = $2, icon = $3, pinned = $4, description = $5, published = $6
		WHERE id = $7
	`, section, name, icon, pinned, description, published, id)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}
