package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"chameth.com/chameth.com/admin/templates"
	"chameth.com/chameth.com/features/wow"
)

func ListWowCharactersHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		characters, err := wow.AllCharacters(r.Context())
		if err != nil {
			slog.Error("Failed to get WoW characters", "error", err)
			http.Error(w, "Failed to retrieve characters", http.StatusInternalServerError)
			return
		}

		summaries := make([]templates.WowCharacterSummary, len(characters))
		for i, c := range characters {
			summaries[i] = templates.WowCharacterSummary{
				ID:            c.ID,
				CharacterName: c.CharacterName,
				RealmName:     c.RealmName,
				Race:          c.Race,
				Class:         c.Class,
				Spec:          c.Spec,
				Gender:        c.Gender,
				Faction:       c.Faction,
				UpdatedAt:     c.UpdatedAt.Format("2006-01-02 15:04"),
			}
		}

		data := templates.ListWowCharactersData{
			Characters: summaries,
		}

		if err := templates.RenderListWowCharacters(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func ImportWowCharacterHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		realm := r.FormValue("realm")
		character := r.FormValue("character")
		if realm == "" || character == "" {
			http.Error(w, "Realm and character are required", http.StatusBadRequest)
			return
		}

		if err := wow.ImportCharacter(r.Context(), realm, character); err != nil {
			slog.Error("Failed to import WoW character", "error", err, "realm", realm, "character", character)
		}

		http.Redirect(w, r, "/wow", http.StatusSeeOther)
	}
}

func RefreshWowCharacterHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid character ID", http.StatusBadRequest)
			return
		}

		if err := wow.RefreshCharacter(r.Context(), id); err != nil {
			slog.Error("Failed to refresh WoW character", "error", err, "id", id)
		}

		http.Redirect(w, r, "/wow", http.StatusSeeOther)
	}
}
