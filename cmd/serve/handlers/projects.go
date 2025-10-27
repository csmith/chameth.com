package handlers

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/csmith/chameth.com/cmd/serve/assets"
	"github.com/csmith/chameth.com/cmd/serve/content"
	"github.com/csmith/chameth.com/cmd/serve/db"
	"github.com/csmith/chameth.com/cmd/serve/templates"
)

func ProjectsList(w http.ResponseWriter, r *http.Request) {
	sections, err := db.GetAllProjectSections()
	if err != nil {
		slog.Error("Failed to get all project sections", "error", err)
		ServerError(w, r)
		return
	}

	var groups []templates.ProjectGroup
	for _, section := range sections {
		var projectDetails []templates.ProjectDetails

		projects, err := db.GetProjectsInSection(section.ID)
		if err != nil {
			slog.Error("Failed to get projects in section", "section", section.ID, "error", err)
			ServerError(w, r)
			return
		}

		for _, project := range projects {
			renderedDesc, err := content.RenderMarkdown(project.Description)
			if err != nil {
				slog.Error("Failed to render markdown for project description", "project", project.Name, "error", err)
				ServerError(w, r)
				return
			}
			projectDetails = append(projectDetails, templates.ProjectDetails{
				Name:        project.Name,
				Pinned:      project.Pinned,
				Icon:        template.HTML(project.Icon),
				Description: renderedDesc,
			})
		}

		groups = append(groups, templates.ProjectGroup{
			Name:        section.Name,
			Description: section.Description,
			Projects:    projectDetails,
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderProjects(w, templates.ProjectsData{
		ProjectGroups: groups,
		PageData: templates.PageData{
			Title:        "Projects · Chameth.com",
			Stylesheet:   assets.GetStylesheetPath(),
			CanonicalUrl: "https://chameth.com/projects/",
			RecentPosts:  content.RecentPosts(),
		},
	})
}
