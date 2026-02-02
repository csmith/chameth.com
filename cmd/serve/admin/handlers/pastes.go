package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"chameth.com/chameth.com/cmd/serve/admin/templates"
	"chameth.com/chameth.com/cmd/serve/db"
	"github.com/csmith/aca"
)

func ListPastesHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		drafts, err := db.GetDraftPastes(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve draft pastes", http.StatusInternalServerError)
			return
		}

		pastes, err := db.GetAllPastes(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve pastes", http.StatusInternalServerError)
			return
		}

		draftSummaries := make([]templates.PasteSummary, len(drafts))
		for i, paste := range drafts {
			draftSummaries[i] = templates.PasteSummary{
				ID:       paste.ID,
				Path:     paste.Path,
				Title:    paste.Title,
				Language: paste.Language,
			}
		}

		pasteSummaries := make([]templates.PasteSummary, len(pastes))
		for i, paste := range pastes {
			pasteSummaries[i] = templates.PasteSummary{
				ID:       paste.ID,
				Path:     paste.Path,
				Title:    paste.Title,
				Language: paste.Language,
			}
		}

		data := templates.ListPastesData{
			Drafts: draftSummaries,
			Pastes: pasteSummaries,
		}

		if err := templates.RenderListPastes(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func EditPasteHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid paste ID", http.StatusBadRequest)
			return
		}

		paste, err := db.GetPasteByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Paste not found", http.StatusNotFound)
			return
		}

		data := templates.EditPasteData{
			ID:        paste.ID,
			Path:      paste.Path,
			Title:     paste.Title,
			Language:  paste.Language,
			Content:   paste.Content,
			Date:      paste.Date.Format("2006-01-02"),
			Published: paste.Published,
		}

		if err := templates.RenderEditPaste(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func CreatePasteHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate random adjective-color-animal name
		gen, err := aca.NewDefaultGenerator()
		if err != nil {
			http.Error(w, "Failed to generate name", http.StatusInternalServerError)
			return
		}
		name := gen.Generate()
		path := fmt.Sprintf("/paste/%s/", name)

		// Create the new paste
		id, err := db.CreatePaste(r.Context(), path, name)
		if err != nil {
			http.Error(w, "Failed to create paste", http.StatusInternalServerError)
			return
		}

		// Redirect to edit page
		http.Redirect(w, r, fmt.Sprintf("/pastes/edit/%d", id), http.StatusSeeOther)
	}
}

func UpdatePasteHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid paste ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		path := r.FormValue("path")
		title := r.FormValue("title")
		language := r.FormValue("language")
		pasteContent := r.FormValue("content")
		date := r.FormValue("date")
		published := r.FormValue("published") == "true"

		if err := db.UpdatePaste(r.Context(), id, path, title, language, pasteContent, date, published); err != nil {
			http.Error(w, "Failed to update paste", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/pastes/edit/%d", id), http.StatusSeeOther)
	}
}
