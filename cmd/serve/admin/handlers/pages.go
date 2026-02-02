package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"chameth.com/chameth.com/cmd/serve/admin/templates"
	"chameth.com/chameth.com/cmd/serve/db"
	"github.com/csmith/aca"
)

func ListPagesHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		drafts, err := db.GetDraftStaticPages(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve draft pages", http.StatusInternalServerError)
			return
		}

		pages, err := db.GetAllStaticPages(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve pages", http.StatusInternalServerError)
			return
		}

		draftSummaries := make([]templates.PageSummary, len(drafts))
		for i, page := range drafts {
			draftSummaries[i] = templates.PageSummary{
				ID:    page.ID,
				Title: page.Title,
				Path:  page.Path,
			}
		}

		pageSummaries := make([]templates.PageSummary, len(pages))
		for i, page := range pages {
			pageSummaries[i] = templates.PageSummary{
				ID:    page.ID,
				Title: page.Title,
				Path:  page.Path,
			}
		}

		data := templates.ListPagesData{
			Drafts: draftSummaries,
			Pages:  pageSummaries,
		}

		if err := templates.RenderListPages(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func EditPageHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid page ID", http.StatusBadRequest)
			return
		}

		page, err := db.GetStaticPageByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Page not found", http.StatusNotFound)
			return
		}

		// Fetch media relations for this page
		mediaRelations, err := db.GetMediaRelationsForEntity(r.Context(), "staticpage", id)
		if err != nil {
			http.Error(w, "Failed to retrieve media", http.StatusInternalServerError)
			return
		}

		// Group media by primary vs variants
		// Use two-pass approach to handle cases where variants appear before their parents
		mediaMap := make(map[int]*templates.PageMediaItem)
		var primaryMediaIDs []int

		// First pass: add all primary media items
		for _, rel := range mediaRelations {
			if rel.ParentMediaID == nil {
				if _, exists := mediaMap[rel.MediaID]; !exists {
					primaryMediaIDs = append(primaryMediaIDs, rel.MediaID)

					caption := ""
					if rel.Caption != nil {
						caption = *rel.Caption
					}
					description := ""
					if rel.Description != nil {
						description = *rel.Description
					}
					role := ""
					if rel.Role != nil {
						role = *rel.Role
					}

					mediaMap[rel.MediaID] = &templates.PageMediaItem{
						Path:        rel.Path,
						Title:       caption,
						AltText:     description,
						Width:       rel.Width,
						Height:      rel.Height,
						Role:        role,
						ContentType: rel.ContentType,
						MediaID:     rel.MediaID,
						Variants:    []templates.PageMediaVariant{},
					}
				}
			}
		}

		// Second pass: add all variants to their parents
		for _, rel := range mediaRelations {
			if rel.ParentMediaID != nil {
				parentID := *rel.ParentMediaID
				if parent, exists := mediaMap[parentID]; exists {
					parent.Variants = append(parent.Variants, templates.PageMediaVariant{
						MediaID:     rel.MediaID,
						ContentType: rel.ContentType,
						Width:       rel.Width,
						Height:      rel.Height,
					})
				}
			}
		}

		// Convert map to slice in order of discovery
		mediaItems := make([]templates.PageMediaItem, 0, len(primaryMediaIDs))
		for _, mediaID := range primaryMediaIDs {
			mediaItems = append(mediaItems, *mediaMap[mediaID])
		}

		data := templates.EditPageData{
			ID:        page.ID,
			Title:     page.Title,
			Path:      page.Path,
			Content:   page.Content,
			Published: page.Published,
			Raw:       page.Raw,
			Media:     mediaItems,
		}

		if err := templates.RenderEditPage(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func CreatePageHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate random adjective-color-animal name
		gen, err := aca.NewDefaultGenerator()
		if err != nil {
			http.Error(w, "Failed to generate name", http.StatusInternalServerError)
			return
		}
		name := gen.Generate()
		path := fmt.Sprintf("/%s/", name)

		// Create the new page
		id, err := db.CreateStaticPage(r.Context(), path, name)
		if err != nil {
			http.Error(w, "Failed to create page", http.StatusInternalServerError)
			return
		}

		// Redirect to edit page
		http.Redirect(w, r, fmt.Sprintf("/pages/edit/%d", id), http.StatusSeeOther)
	}
}

func UpdatePageHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid page ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		path := r.FormValue("path")
		title := r.FormValue("title")
		pageContent := r.FormValue("content")
		published := r.FormValue("published") == "true"
		raw := r.FormValue("raw") == "true"

		if err := db.UpdateStaticPage(r.Context(), id, path, title, pageContent, published, raw); err != nil {
			http.Error(w, "Failed to update page", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/pages/edit/%d", id), http.StatusSeeOther)
	}
}
