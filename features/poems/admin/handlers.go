package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"chameth.com/chameth.com/features/poems"
	"chameth.com/chameth.com/features/poems/admin/templates"
	"github.com/csmith/aca"
)

func ListPoemsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		drafts, err := poems.GetDraftPoems(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve draft poems", http.StatusInternalServerError)
			return
		}

		allPoems, err := poems.GetAllPoems(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve poems", http.StatusInternalServerError)
			return
		}

		draftSummaries := make([]templates.PoemSummary, len(drafts))
		for i, poem := range drafts {
			draftSummaries[i] = templates.PoemSummary{
				ID:    poem.ID,
				Path:  poem.Path,
				Title: poem.Title,
				Date:  poem.Date.Format("2006-01-02"),
			}
		}

		poemSummaries := make([]templates.PoemSummary, len(allPoems))
		for i, poem := range allPoems {
			poemSummaries[i] = templates.PoemSummary{
				ID:    poem.ID,
				Path:  poem.Path,
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

		poem, err := poems.GetPoemByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Poem not found", http.StatusNotFound)
			return
		}

		data := templates.EditPoemData{
			ID:        poem.ID,
			Path:      poem.Path,
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
		gen, err := aca.NewDefaultGenerator()
		if err != nil {
			http.Error(w, "Failed to generate name", http.StatusInternalServerError)
			return
		}
		name := gen.Generate()
		path := fmt.Sprintf("/%s/", name)

		id, err := poems.CreatePoem(r.Context(), path, name)
		if err != nil {
			http.Error(w, "Failed to create poem", http.StatusInternalServerError)
			return
		}

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

		path := r.FormValue("path")
		title := r.FormValue("title")
		poemContent := r.FormValue("poem")
		notes := r.FormValue("notes")
		date := r.FormValue("date")
		published := r.FormValue("published") == "true"

		if err := poems.UpdatePoem(r.Context(), id, path, title, poemContent, notes, date, published); err != nil {
			http.Error(w, "Failed to update poem", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/poems/edit/%d", id), http.StatusSeeOther)
	}
}
