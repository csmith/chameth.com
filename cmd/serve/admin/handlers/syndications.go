package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"chameth.com/chameth.com/cmd/serve/admin/templates"
	"chameth.com/chameth.com/cmd/serve/db"
)

func ListSyndicationsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		unpublished, err := db.GetUnpublishedSyndications()
		if err != nil {
			http.Error(w, "Failed to retrieve unpublished syndications", http.StatusInternalServerError)
			return
		}

		syndications, err := db.GetAllSyndications()
		if err != nil {
			http.Error(w, "Failed to retrieve syndications", http.StatusInternalServerError)
			return
		}

		unpublishedSummaries := make([]templates.SyndicationSummary, len(unpublished))
		for i, s := range unpublished {
			unpublishedSummaries[i] = templates.SyndicationSummary{
				ID:          s.ID,
				Path:        s.Path,
				ExternalURL: s.ExternalURL,
				Name:        s.Name,
				Published:   s.Published,
			}
		}

		syndicationSummaries := make([]templates.SyndicationSummary, len(syndications))
		for i, s := range syndications {
			syndicationSummaries[i] = templates.SyndicationSummary{
				ID:          s.ID,
				Path:        s.Path,
				ExternalURL: s.ExternalURL,
				Name:        s.Name,
				Published:   s.Published,
			}
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

		syndication, err := db.GetSyndicationByID(id)
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

		if path == "" || externalURL == "" || name == "" {
			http.Error(w, "Path, external URL, and name are required", http.StatusBadRequest)
			return
		}

		id, err := db.CreateSyndication(path, externalURL, name, false)
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

		if err := db.UpdateSyndication(id, path, externalURL, name, published); err != nil {
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

		if err := db.DeleteSyndication(id); err != nil {
			http.Error(w, "Failed to delete syndication", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/syndications", http.StatusSeeOther)
	}
}
