package db

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
