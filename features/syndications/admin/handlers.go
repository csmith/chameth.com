package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"chameth.com/chameth.com/features/syndications"
	"chameth.com/chameth.com/features/syndications/admin/templates"
)

func ListSyndicationsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		unpublished, err := syndications.GetUnpublishedSyndications(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve unpublished syndications", http.StatusInternalServerError)
			return
		}

		all, err := syndications.GetAllSyndications(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve syndications", http.StatusInternalServerError)
			return
		}

		unpublishedSummaries := make([]templates.SyndicationSummary, len(unpublished))
		for i, s := range unpublished {
			summary := templates.SyndicationSummary{
				ID:          s.ID,
				Path:        s.Path,
				ExternalURL: s.ExternalURL,
				Name:        s.Name,
				Published:   s.Published,
				Disposition: s.Disposition,
			}
			if s.Rel != nil {
				summary.Rel = *s.Rel
			}
			unpublishedSummaries[i] = summary
		}

		syndicationSummaries := make([]templates.SyndicationSummary, len(all))
		for i, s := range all {
			summary := templates.SyndicationSummary{
				ID:          s.ID,
				Path:        s.Path,
				ExternalURL: s.ExternalURL,
				Name:        s.Name,
				Published:   s.Published,
				Disposition: s.Disposition,
			}
			if s.Rel != nil {
				summary.Rel = *s.Rel
			}
			syndicationSummaries[i] = summary
		}

		data := templates.ListSyndicationsData{
			Unpublished:  unpublishedSummaries,
			Syndications: syndicationSummaries,
		}

		if err := templates.RenderListSyndications(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func EditSyndicationHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid syndication ID", http.StatusBadRequest)
			return
		}

		syndication, err := syndications.GetSyndicationByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Syndication not found", http.StatusNotFound)
			return
		}

		data := templates.EditSyndicationData{
			ID:          syndication.ID,
			Path:        syndication.Path,
			ExternalURL: syndication.ExternalURL,
			Name:        syndication.Name,
			Published:   syndication.Published,
			Disposition: syndication.Disposition,
		}
		if syndication.Rel != nil {
			data.Rel = *syndication.Rel
		}

		if err := templates.RenderEditSyndication(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func CreateSyndicationHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		path := r.FormValue("path")
		externalURL := r.FormValue("external_url")
		name := r.FormValue("name")
		disposition := r.FormValue("disposition")
		relStr := r.FormValue("rel")

		if path == "" || externalURL == "" || name == "" {
			http.Error(w, "Path, external URL, and name are required", http.StatusBadRequest)
			return
		}

		if disposition == "" {
			disposition = "anchor"
		}

		var rel *string
		if relStr != "" {
			rel = &relStr
		}

		id, err := syndications.CreateSyndication(r.Context(), path, externalURL, name, false, disposition, rel)
		if err != nil {
			http.Error(w, "Failed to create syndication", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/syndications/edit/%d", id), http.StatusSeeOther)
	}
}

func UpdateSyndicationHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid syndication ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		path := r.FormValue("path")
		externalURL := r.FormValue("external_url")
		name := r.FormValue("name")
		published := r.FormValue("published") == "true"
		disposition := r.FormValue("disposition")
		relStr := r.FormValue("rel")

		var rel *string
		if relStr != "" {
			rel = &relStr
		}

		if err := syndications.UpdateSyndication(r.Context(), id, path, externalURL, name, published, disposition, rel); err != nil {
			http.Error(w, "Failed to update syndication", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/syndications/edit/%d", id), http.StatusSeeOther)
	}
}

func DeleteSyndicationHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid syndication ID", http.StatusBadRequest)
			return
		}

		if err := syndications.DeleteSyndication(r.Context(), id); err != nil {
			http.Error(w, "Failed to delete syndication", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/syndications", http.StatusSeeOther)
	}
}
