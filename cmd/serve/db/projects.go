package db

import "fmt"

// GetAllProjectSections returns all project sections ordered by sort.
func GetAllProjectSections() ([]ProjectSection, error) {
	var sections []ProjectSection
	err := db.Select(&sections, "SELECT id, name, sort, description FROM project_sections ORDER BY sort")
	if err != nil {
		return nil, err
	}
	return sections, nil
}

// GetProjectsInSection returns all projects in a section ordered by pinned (descending), then name (case-insensitive).
func GetProjectsInSection(sectionID int) ([]Project, error) {
	var projects []Project
	err := db.Select(&projects, "SELECT id, section, name, icon, pinned, description FROM projects WHERE section = $1 AND published = true ORDER BY pinned DESC, LOWER(name)", sectionID)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// GetAllProjects returns all published projects.
func GetAllProjects() ([]Project, error) {
	var projects []Project
	err := db.Select(&projects, "SELECT id, section, name, icon, pinned, description FROM projects WHERE published = true ORDER BY section, pinned DESC, LOWER(name)")
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// GetDraftProjects returns all unpublished projects.
func GetDraftProjects() ([]Project, error) {
	var projects []Project
	err := db.Select(&projects, "SELECT id, section, name, icon, pinned, description FROM projects WHERE published = false ORDER BY section, pinned DESC, LOWER(name)")
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProjectByID returns a project for the given ID.
func GetProjectByID(id int) (*Project, error) {
	var project Project
	err := db.Get(&project, "SELECT id, section, name, icon, pinned, description, published FROM projects WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// CreateProject creates a new project in the database and returns its ID.
// The project is created with the first section by sort order, unpublished, not pinned.
func CreateProject(name string) (int, error) {
	sections, err := GetAllProjectSections()
	if err != nil {
		return 0, fmt.Errorf("failed to get sections: %w", err)
	}
	if len(sections) == 0 {
		return 0, fmt.Errorf("no sections available")
	}
	defaultSection := sections[0].ID

	var id int
	err = db.QueryRow(`
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
func UpdateProject(id int, name, icon, description string, section int, pinned, published bool) error {
	_, err := db.Exec(`
		UPDATE projects
		SET section = $1, name = $2, icon = $3, pinned = $4, description = $5, published = $6
		WHERE id = $7
	`, section, name, icon, pinned, description, published, id)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}
