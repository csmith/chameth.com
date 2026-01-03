package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/csmith/chameth.com/cmd/serve/admin/templates"
	"github.com/csmith/chameth.com/cmd/serve/db"
)

func ListGoImportsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		drafts, err := db.GetDraftGoImports()
		if err != nil {
			http.Error(w, "Failed to retrieve draft goimports", http.StatusInternalServerError)
			return
		}

		goimports, err := db.GetAllGoImports()
		if err != nil {
			http.Error(w, "Failed to retrieve goimports", http.StatusInternalServerError)
			return
		}

		draftSummaries := make([]templates.GoImportSummary, len(drafts))
		for i, gi := range drafts {
			draftSummaries[i] = templates.GoImportSummary{
				ID:      gi.ID,
				Path:    gi.Path,
				VCS:     gi.VCS,
				RepoURL: gi.RepoURL,
			}
		}

		goimportSummaries := make([]templates.GoImportSummary, len(goimports))
		for i, gi := range goimports {
			goimportSummaries[i] = templates.GoImportSummary{
				ID:      gi.ID,
				Path:    gi.Path,
				VCS:     gi.VCS,
				RepoURL: gi.RepoURL,
			}
		}

		data := templates.ListGoImportsData{
			Drafts:    draftSummaries,
			GoImports: goimportSummaries,
		}

		if err := templates.RenderListGoImports(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func EditGoImportHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid goimport ID", http.StatusBadRequest)
			return
		}

		goimport, err := db.GetGoImportByID(id)
		if err != nil {
			http.Error(w, "Go import not found", http.StatusNotFound)
			return
		}

		data := templates.EditGoImportData{
			ID:        goimport.ID,
			Path:      goimport.Path,
			VCS:       goimport.VCS,
			RepoURL:   goimport.RepoURL,
			Published: goimport.Published,
		}

		if err := templates.RenderEditGoImport(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func CreateGoImportHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		project := r.FormValue("project")
		if project == "" {
			http.Error(w, "Project name is required", http.StatusBadRequest)
			return
		}

		path := "/" + project + "/"
		vcs := "git"
		repoUrl := "https://github.com/csmith/" + project

		id, err := db.CreateGoImport(path, vcs, repoUrl)
		if err != nil {
			http.Error(w, "Failed to create goimport", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/goimports/edit/%d", id), http.StatusSeeOther)
	}
}

func UpdateGoImportHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid goimport ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		path := r.FormValue("path")
		vcs := r.FormValue("vcs")
		repoUrl := r.FormValue("repo_url")
		published := r.FormValue("published") == "true"

		if err := db.UpdateGoImport(id, path, vcs, repoUrl, published); err != nil {
			http.Error(w, "Failed to update goimport", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/goimports/edit/%d", id), http.StatusSeeOther)
	}
}
