package projects

import (
	"html/template"
	"log/slog"
	"net/http"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/content/markdown"
	projecttemplates "chameth.com/chameth.com/features/projects/templates"
	parenttemplates "chameth.com/chameth.com/templates"
)

func handleList(w http.ResponseWriter, r *http.Request) {
	sections, err := GetAllProjectSections(r.Context())
	if err != nil {
		slog.Error("Failed to get all project sections", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var groups []projecttemplates.ProjectGroup
	for _, section := range sections {
		var projectDetails []projecttemplates.ProjectDetails

		projects, err := GetProjectsInSection(r.Context(), section.ID)
		if err != nil {
			slog.Error("Failed to get projects in section", "section", section.ID, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		for _, project := range projects {
			renderedDesc, err := markdown.Render(project.Description)
			if err != nil {
				slog.Error("Failed to render markdown for project description", "project", project.Name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			projectDetails = append(projectDetails, projecttemplates.ProjectDetails{
				Name:        project.Name,
				Pinned:      project.Pinned,
				Icon:        template.HTML(project.Icon),
				Description: renderedDesc,
			})
		}

		groups = append(groups, projecttemplates.ProjectGroup{
			Name:        section.Name,
			Description: section.Description,
			Projects:    projectDetails,
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = projecttemplates.RenderProjects(w, projecttemplates.ProjectsData{
		ProjectGroups: groups,
		PageData:      content.CreatePageData(r.Context(), "Projects", "/projects/", parenttemplates.OpenGraphHeaders{}),
	})
}
