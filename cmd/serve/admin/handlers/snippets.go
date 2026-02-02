package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"chameth.com/chameth.com/cmd/serve/admin/templates"
	"chameth.com/chameth.com/cmd/serve/db"
	"github.com/csmith/aca"
)

func ListSnippetsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		drafts, err := db.GetDraftSnippets(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve draft snippets", http.StatusInternalServerError)
			return
		}

		snippets, err := db.GetAllSnippets(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve snippets", http.StatusInternalServerError)
			return
		}

		draftSummaries := make([]templates.SnippetSummary, len(drafts))
		for i, snippet := range drafts {
			draftSummaries[i] = templates.SnippetSummary{
				ID:    snippet.ID,
				Path:  snippet.Path,
				Title: snippet.Title,
				Topic: snippet.Topic,
			}
		}

		snippetSummaries := make([]templates.SnippetSummary, len(snippets))
		for i, snippet := range snippets {
			snippetSummaries[i] = templates.SnippetSummary{
				ID:    snippet.ID,
				Path:  snippet.Path,
				Title: snippet.Title,
				Topic: snippet.Topic,
			}
		}

		data := templates.ListSnippetsData{
			Drafts:   draftSummaries,
			Snippets: snippetSummaries,
		}

		if err := templates.RenderListSnippets(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func EditSnippetHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid snippet ID", http.StatusBadRequest)
			return
		}

		snippet, err := db.GetSnippetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Snippet not found", http.StatusNotFound)
			return
		}

		topics, err := db.GetAllTopics(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve topics", http.StatusInternalServerError)
			return
		}

		data := templates.EditSnippetData{
			ID:              snippet.ID,
			Path:            snippet.Path,
			Title:           snippet.Title,
			Topic:           snippet.Topic,
			Content:         snippet.Content,
			Published:       snippet.Published,
			AvailableTopics: topics,
		}

		if err := templates.RenderEditSnippet(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func CreateSnippetHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate random adjective-color-animal name
		gen, err := aca.NewDefaultGenerator()
		if err != nil {
			http.Error(w, "Failed to generate name", http.StatusInternalServerError)
			return
		}
		name := gen.Generate()
		path := fmt.Sprintf("/snippets/%s/", name)

		// Create the new snippet
		id, err := db.CreateSnippet(r.Context(), path, name)
		if err != nil {
			http.Error(w, "Failed to create snippet", http.StatusInternalServerError)
			return
		}

		// Redirect to edit page
		http.Redirect(w, r, fmt.Sprintf("/snippets/edit/%d", id), http.StatusSeeOther)
	}
}

func UpdateSnippetHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid snippet ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		path := r.FormValue("path")
		title := r.FormValue("title")
		snippetContent := r.FormValue("content")
		published := r.FormValue("published") == "true"

		// Use custom topic if provided, otherwise use selected topic
		topic := r.FormValue("custom_topic")
		if topic == "" {
			topic = r.FormValue("topic")
		}

		if err := db.UpdateSnippet(r.Context(), id, path, title, topic, snippetContent, published); err != nil {
			http.Error(w, "Failed to update snippet", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/snippets/edit/%d", id), http.StatusSeeOther)
	}
}
