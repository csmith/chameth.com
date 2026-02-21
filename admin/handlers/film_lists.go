package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"chameth.com/chameth.com/admin/templates"
	"chameth.com/chameth.com/db"
	"github.com/csmith/aca"
)

func ListFilmListsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		drafts, err := db.GetDraftFilmLists(r.Context())
		if err != nil {
			slog.Error("Failed to retrieve draft film lists", "error", err)
			http.Error(w, "Failed to retrieve draft film lists", http.StatusInternalServerError)
			return
		}

		lists, err := db.GetAllFilmLists(r.Context())
		if err != nil {
			slog.Error("Failed to retrieve film lists", "error", err)
			http.Error(w, "Failed to retrieve film lists", http.StatusInternalServerError)
			return
		}

		draftSummaries := make([]templates.FilmListSummary, len(drafts))
		for i, list := range drafts {
			draftSummaries[i] = templates.FilmListSummary{
				ID:        list.ID,
				Title:     list.Title,
				Path:      list.Path,
				Published: list.Published,
			}
		}

		listSummaries := make([]templates.FilmListSummary, len(lists))
		for i, list := range lists {
			listSummaries[i] = templates.FilmListSummary{
				ID:        list.ID,
				Title:     list.Title,
				Path:      list.Path,
				Published: list.Published,
			}
		}

		data := templates.ListFilmListsData{
			Drafts: draftSummaries,
			Lists:  listSummaries,
		}

		if err := templates.RenderListFilmLists(w, data); err != nil {
			slog.Error("Failed to render film lists template", "error", err)
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func CreateFilmListHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		gen, err := aca.NewDefaultGenerator()
		if err != nil {
			slog.Error("Failed to create name generator for film list", "error", err)
			http.Error(w, "Failed to create film list", http.StatusInternalServerError)
			return
		}
		name := gen.Generate()
		path := fmt.Sprintf("/films/lists/%s/", name)

		id, err := db.CreateFilmList(r.Context(), path, name, "")
		if err != nil {
			slog.Error("Failed to create film list", "error", err)
			http.Error(w, "Failed to create film list", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/film-lists/%d/edit", id), http.StatusSeeOther)
	}
}

func EditFilmListHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid film list ID", http.StatusBadRequest)
			return
		}

		list, entries, err := db.GetFilmListWithEntries(r.Context(), id)
		if err != nil {
			slog.Error("Failed to get film list with entries", "id", id, "error", err)
			http.Error(w, "Film list not found", http.StatusNotFound)
			return
		}

		allFilms, err := db.GetAllFilms(r.Context())
		if err != nil {
			slog.Error("Failed to get all films", "error", err)
			http.Error(w, "Failed to get films", http.StatusInternalServerError)
			return
		}

		entryItems := make([]templates.FilmListEntryItem, len(entries))
		existingFilmIDs := make(map[int]struct{}, len(entries))
		for i, entry := range entries {
			year := ""
			if entry.Film.Year != nil {
				year = strconv.Itoa(*entry.Film.Year)
			}

			entryItems[i] = templates.FilmListEntryItem{
				EntryID:  entry.ID,
				FilmID:   entry.Film.ID,
				Position: entry.Position,
				Title:    entry.Film.Title,
				Year:     year,
			}
			existingFilmIDs[entry.Film.ID] = struct{}{}
		}

		availableFilms := make([]db.Film, 0, len(allFilms))
		for _, film := range allFilms {
			if _, exists := existingFilmIDs[film.ID]; !exists {
				availableFilms = append(availableFilms, film)
			}
		}

		filmOptions := make([]templates.FilmOption, len(availableFilms))
		for i, film := range availableFilms {
			year := ""
			if film.Year != nil {
				year = strconv.Itoa(*film.Year)
			}
			filmOptions[i] = templates.FilmOption{
				ID:    film.ID,
				Title: film.Title,
				Year:  year,
			}
		}

		data := templates.EditFilmListData{
			ID:          list.ID,
			Title:       list.Title,
			Description: list.Description,
			Published:   list.Published,
			Path:        list.Path,
			Entries:     entryItems,
			Films:       filmOptions,
		}

		if err := templates.RenderEditFilmList(w, data); err != nil {
			slog.Error("Failed to render edit film list template", "id", id, "error", err)
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func UpdateFilmListHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			slog.Error("Invalid film list ID", "id_str", idStr, "error", err)
			http.Error(w, "Invalid film list ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			slog.Error("Failed to parse form when updating film list", "id", id, "error", err)
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		path := r.FormValue("path")
		title := r.FormValue("title")
		description := r.FormValue("description")
		published := r.FormValue("published") == "true"

		if err := db.UpdateFilmList(r.Context(), id, path, title, description, published); err != nil {
			slog.Error("Failed to update film list", "id", id, "error", err)
			http.Error(w, "Failed to update film list", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/film-lists/%d/edit", id), http.StatusSeeOther)
	}
}

func AddFilmToListHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			slog.Error("Invalid film list ID", "id_str", idStr, "error", err)
			http.Error(w, "Invalid film list ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			slog.Error("Failed to parse form when adding film to list", "list_id", id, "error", err)
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		filmIDStr := r.FormValue("film_id")
		if filmIDStr == "" {
			http.Redirect(w, r, fmt.Sprintf("/film-lists/%d/edit", id), http.StatusSeeOther)
			return
		}

		filmID, err := strconv.Atoi(filmIDStr)
		if err != nil {
			slog.Error("Invalid film ID", "film_id_str", filmIDStr, "error", err)
			http.Error(w, "Invalid film ID", http.StatusBadRequest)
			return
		}

		position, err := db.GetNextPosition(r.Context(), id)
		if err != nil {
			slog.Error("Failed to get next position for film list", "id", id, "error", err)
			http.Error(w, "Failed to get next position", http.StatusInternalServerError)
			return
		}

		_, err = db.AddFilmToList(r.Context(), id, filmID, position)
		if err != nil {
			slog.Error("Failed to add film to list", "list_id", id, "film_id", filmID, "error", err)
			http.Error(w, "Failed to add film to list", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/film-lists/%d/edit", id), http.StatusSeeOther)
	}
}

func RemoveFilmFromListHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid film list ID", http.StatusBadRequest)
			return
		}

		entryIDStr := r.PathValue("entryId")
		entryID, err := strconv.Atoi(entryIDStr)
		if err != nil {
			http.Error(w, "Invalid entry ID", http.StatusBadRequest)
			return
		}

		if err := db.RemoveFilmFromList(r.Context(), entryID); err != nil {
			slog.Error("Failed to remove film from list", "list_id", id, "entry_id", entryID, "error", err)
			http.Error(w, "Failed to remove film from list", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/film-lists/%d/edit", id), http.StatusSeeOther)
	}
}

func UpdateEntryPositionHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			slog.Error("Invalid film list ID", "id_str", idStr, "error", err)
			http.Error(w, "Invalid film list ID", http.StatusBadRequest)
			return
		}

		entryIDStr := r.PathValue("entryId")
		entryID, err := strconv.Atoi(entryIDStr)
		if err != nil {
			slog.Error("Invalid entry ID", "entry_id_str", entryIDStr, "error", err)
			http.Error(w, "Invalid entry ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			slog.Error("Failed to parse form when updating position", "list_id", id, "entry_id", entryID, "error", err)
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		positionStr := r.FormValue("position")
		position, err := strconv.Atoi(positionStr)
		if err != nil {
			slog.Error("Invalid position", "position_str", positionStr, "error", err)
			http.Error(w, "Invalid position", http.StatusBadRequest)
			return
		}

		if err := db.UpdateEntryPosition(r.Context(), entryID, position); err != nil {
			slog.Error("Failed to update entry position", "entry_id", entryID, "new_position", position, "error", err)
			http.Error(w, "Failed to update position", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/film-lists/%d/edit", id), http.StatusSeeOther)
	}
}

func ReorderFilmListEntriesHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			slog.Error("Invalid film list ID", "id_str", idStr, "error", err)
			http.Error(w, "Invalid film list ID", http.StatusBadRequest)
			return
		}

		if err := db.ReorderFilmListEntries(r.Context(), id); err != nil {
			slog.Error("Failed to reorder film list entries", "id", id, "error", err)
			http.Error(w, "Failed to reorder entries", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/film-lists/%d/edit", id), http.StatusSeeOther)
	}
}
