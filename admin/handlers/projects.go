package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"chameth.com/chameth.com/admin/templates"
	"chameth.com/chameth.com/db"
	"github.com/csmith/aca"
)

func ListProjectsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		drafts, err := db.GetDraftProjects(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve draft projects", http.StatusInternalServerError)
			return
		}

		projects, err := db.GetAllProjects(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve projects", http.StatusInternalServerError)
			return
		}

		sections, err := db.GetAllProjectSections(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve sections", http.StatusInternalServerError)
			return
		}

		sectionMap := make(map[int]string)
		for _, section := range sections {
			sectionMap[section.ID] = section.Name
		}

		sectionOrderMap := make(map[int]int)
		for i, section := range sections {
			sectionOrderMap[section.ID] = i
		}

		draftSummaries := make([]templates.ProjectSummary, len(drafts))
		for i, project := range drafts {
			draftSummaries[i] = templates.ProjectSummary{
				ID:          project.ID,
				Name:        project.Name,
				Icon:        template.HTML(project.Icon),
				Pinned:      project.Pinned,
				Section:     sectionMap[project.Section],
				Description: project.Description,
				SectionSort: sectionOrderMap[project.Section],
			}
		}

		projectSummaries := make([]templates.ProjectSummary, len(projects))
		for i, project := range projects {
			projectSummaries[i] = templates.ProjectSummary{
				ID:          project.ID,
				Name:        project.Name,
				Icon:        template.HTML(project.Icon),
				Pinned:      project.Pinned,
				Section:     sectionMap[project.Section],
				Description: project.Description,
				SectionSort: sectionOrderMap[project.Section],
			}
		}

		data := templates.ListProjectsData{
			Drafts:   draftSummaries,
			Projects: projectSummaries,
		}

		if err := templates.RenderListProjects(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func CreateProjectHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		gen, err := aca.NewDefaultGenerator()
		if err != nil {
			http.Error(w, "Failed to generate name", http.StatusInternalServerError)
			return
		}
		name := gen.Generate()

		id, err := db.CreateProject(r.Context(), name)
		if err != nil {
			http.Error(w, "Failed to create project", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/projects/edit/%d", id), http.StatusSeeOther)
	}
}

func EditProjectHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		project, err := db.GetProjectByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		sections, err := db.GetAllProjectSections(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve sections", http.StatusInternalServerError)
			return
		}

		sectionOptions := make([]templates.SectionOption, len(sections))
		for i, section := range sections {
			sectionOptions[i] = templates.SectionOption{
				ID:   section.ID,
				Name: section.Name,
			}
		}

		data := templates.EditProjectData{
			ID:                project.ID,
			Name:              project.Name,
			Icon:              project.Icon,
			Description:       project.Description,
			Section:           project.Section,
			Pinned:            project.Pinned,
			Published:         project.Published,
			AvailableSections: sectionOptions,
		}

		if err := templates.RenderEditProject(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func UpdateProjectHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		icon := r.FormValue("icon")
		description := r.FormValue("description")
		sectionStr := r.FormValue("section")
		section, err := strconv.Atoi(sectionStr)
		if err != nil {
			http.Error(w, "Invalid section ID", http.StatusBadRequest)
			return
		}
		pinned := r.FormValue("pinned") == "true"
		published := r.FormValue("published") == "true"

		if err := db.UpdateProject(r.Context(), id, name, icon, description, section, pinned, published); err != nil {
			http.Error(w, "Failed to update project", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/projects/edit/%d", id), http.StatusSeeOther)
	}
}
