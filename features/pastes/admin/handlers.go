package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"chameth.com/chameth.com/features/pastes"
	"chameth.com/chameth.com/features/pastes/admin/templates"
	"github.com/csmith/aca"
)

func ListPastesHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		drafts, err := pastes.GetDraftPastes(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve draft pastes", http.StatusInternalServerError)
			return
		}

		allPastes, err := pastes.GetAllPastes(r.Context())
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

		pasteSummaries := make([]templates.PasteSummary, len(allPastes))
		for i, paste := range allPastes {
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

		paste, err := pastes.GetPasteByID(r.Context(), id)
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
		gen, err := aca.NewDefaultGenerator()
		if err != nil {
			http.Error(w, "Failed to generate name", http.StatusInternalServerError)
			return
		}
		name := gen.Generate()
		path := fmt.Sprintf("/paste/%s/", name)

		id, err := pastes.CreatePaste(r.Context(), path, name)
		if err != nil {
			http.Error(w, "Failed to create paste", http.StatusInternalServerError)
			return
		}

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

		if err := pastes.UpdatePaste(r.Context(), id, path, title, language, pasteContent, date, published); err != nil {
			http.Error(w, "Failed to update paste", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/pastes/edit/%d", id), http.StatusSeeOther)
	}
}
