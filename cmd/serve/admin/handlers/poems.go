package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/csmith/aca"
	"github.com/csmith/chameth.com/cmd/serve/admin/templates"
	"github.com/csmith/chameth.com/cmd/serve/db"
)

func ListPoemsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		drafts, err := db.GetDraftPoems()
		if err != nil {
			http.Error(w, "Failed to retrieve draft poems", http.StatusInternalServerError)
			return
		}

		poems, err := db.GetAllPoems()
		if err != nil {
			http.Error(w, "Failed to retrieve poems", http.StatusInternalServerError)
			return
		}

		draftSummaries := make([]templates.PoemSummary, len(drafts))
		for i, poem := range drafts {
			draftSummaries[i] = templates.PoemSummary{
				ID:    poem.ID,
				Slug:  poem.Slug,
				Title: poem.Title,
				Date:  poem.Date.Format("2006-01-02"),
			}
		}

		poemSummaries := make([]templates.PoemSummary, len(poems))
		for i, poem := range poems {
			poemSummaries[i] = templates.PoemSummary{
				ID:    poem.ID,
				Slug:  poem.Slug,
				Title: poem.Title,
				Date:  poem.Date.Format("2006-01-02"),
			}
		}

		data := templates.ListPoemsData{
			Drafts: draftSummaries,
			Poems:  poemSummaries,
		}

		if err := templates.RenderListPoems(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func EditPoemHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid poem ID", http.StatusBadRequest)
			return
		}

		poem, err := db.GetPoemByID(id)
		if err != nil {
			http.Error(w, "Poem not found", http.StatusNotFound)
			return
		}

		data := templates.EditPoemData{
			ID:        poem.ID,
			Slug:      poem.Slug,
			Title:     poem.Title,
			Poem:      poem.Poem,
			Notes:     poem.Notes,
			Date:      poem.Date.Format("2006-01-02"),
			Published: poem.Published,
		}

		if err := templates.RenderEditPoem(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func CreatePoemHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate random adjective-color-animal name
		gen, err := aca.NewDefaultGenerator()
		if err != nil {
			http.Error(w, "Failed to generate name", http.StatusInternalServerError)
			return
		}
		name := gen.Generate()
		slug := fmt.Sprintf("/%s/", name)

		// Create the new poem
		id, err := db.CreatePoem(slug, name)
		if err != nil {
			http.Error(w, "Failed to create poem", http.StatusInternalServerError)
			return
		}

		// Redirect to edit page
		http.Redirect(w, r, fmt.Sprintf("/poems/edit/%d", id), http.StatusSeeOther)
	}
}

func UpdatePoemHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid poem ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		slug := r.FormValue("slug")
		title := r.FormValue("title")
		poemContent := r.FormValue("poem")
		notes := r.FormValue("notes")
		date := r.FormValue("date")
		published := r.FormValue("published") == "true"

		if err := db.UpdatePoem(id, slug, title, poemContent, notes, date, published); err != nil {
			http.Error(w, "Failed to update poem", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/poems/edit/%d", id), http.StatusSeeOther)
	}
}
