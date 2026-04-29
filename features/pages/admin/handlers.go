package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/features/pages"
	"chameth.com/chameth.com/features/pages/admin/templates"
	"github.com/csmith/aca"
)

func ListPagesHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		drafts, err := pages.GetDraftStaticPages(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve draft pages", http.StatusInternalServerError)
			return
		}

		allPages, err := pages.GetAllStaticPages(r.Context())
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

		pageSummaries := make([]templates.PageSummary, len(allPages))
		for i, page := range allPages {
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

		page, err := pages.GetStaticPageByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Page not found", http.StatusNotFound)
			return
		}

		mediaRelations, err := db.GetMediaRelationsForEntity(r.Context(), "staticpage", id)
		if err != nil {
			http.Error(w, "Failed to retrieve media", http.StatusInternalServerError)
			return
		}

		mediaMap := make(map[int]*templates.PageMediaItem)
		var primaryMediaIDs []int

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

		mediaItems := make([]templates.PageMediaItem, 0, len(primaryMediaIDs))
		for _, mediaID := range primaryMediaIDs {
			mediaItems = append(mediaItems, *mediaMap[mediaID])
		}

		sitemapFrequency := ""
		if page.SitemapFrequency != nil {
			sitemapFrequency = *page.SitemapFrequency
		}
		sitemapPriority := ""
		if page.SitemapPriority != nil {
			sitemapPriority = fmt.Sprintf("%.1f", *page.SitemapPriority)
		}

		data := templates.EditPageData{
			ID:               page.ID,
			Title:            page.Title,
			Path:             page.Path,
			Content:          page.Content,
			Published:        page.Published,
			Raw:              page.Raw,
			SitemapFrequency: sitemapFrequency,
			SitemapPriority:  sitemapPriority,
			Media:            mediaItems,
		}

		if err := templates.RenderEditPage(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func CreatePageHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		gen, err := aca.NewDefaultGenerator()
		if err != nil {
			http.Error(w, "Failed to generate name", http.StatusInternalServerError)
			return
		}
		name := gen.Generate()
		path := fmt.Sprintf("/%s/", name)

		id, err := pages.CreateStaticPage(r.Context(), path, name)
		if err != nil {
			http.Error(w, "Failed to create page", http.StatusInternalServerError)
			return
		}

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

		var sitemapFrequency *string
		if v := strings.TrimSpace(r.FormValue("sitemap_frequency")); v != "" {
			sitemapFrequency = &v
		}
		var sitemapPriority *float64
		if v := strings.TrimSpace(r.FormValue("sitemap_priority")); v != "" {
			if p, err := strconv.ParseFloat(v, 64); err == nil {
				sitemapPriority = &p
			}
		}

		if err := pages.UpdateStaticPage(r.Context(), id, path, title, pageContent, published, raw, sitemapFrequency, sitemapPriority); err != nil {
			http.Error(w, "Failed to update page", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/pages/edit/%d", id), http.StatusSeeOther)
	}
}
