package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"chameth.com/chameth.com/db"
)

func ImportBoardgamesHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var data struct {
			Games []struct {
				ID      int    `json:"id"`
				UUID    string `json:"uuid"`
				BggID   int    `json:"bggId"`
				BggName string `json:"bggName"`
				BggYear int    `json:"bggYear"`
				Copies  []struct {
					StatusOwned     int `json:"statusOwned"`
					StatusPrevOwned int `json:"statusPrevOwned"`
				} `json:"copies"`
			} `json:"games"`
			Plays []struct {
				UUID      string `json:"uuid"`
				GameRefID int    `json:"gameRefId"`
				PlayDate  string `json:"playDate"`
			} `json:"plays"`
		}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			slog.Error("Failed to decode bgstats data", "error", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Build map of bgstats integer game ID → game UUID for play lookups
		gameIDToUUID := make(map[int]string, len(data.Games))

		for _, g := range data.Games {
			gameIDToUUID[g.ID] = g.UUID

			status := "unowned"
			owned := false
			prevOwned := false
			for _, c := range g.Copies {
				if c.StatusOwned == 1 {
					owned = true
				}
				if c.StatusPrevOwned == 1 {
					prevOwned = true
				}
			}
			if owned {
				status = "owned"
			} else if prevOwned {
				status = "sold"
			}

			err := db.UpsertBoardgameGame(r.Context(), db.BoardgameGame{
				ID:     g.UUID,
				BggID:  g.BggID,
				Name:   g.BggName,
				Year:   g.BggYear,
				Status: status,
			})
			if err != nil {
				slog.Error("Failed to upsert boardgame game", "uuid", g.UUID, "name", g.BggName, "error", err)
				continue
			}
		}

		for _, p := range data.Plays {
			gameUUID, ok := gameIDToUUID[p.GameRefID]
			if !ok {
				slog.Error("Failed to find game for play", "play_uuid", p.UUID, "game_ref_id", p.GameRefID)
				continue
			}

			playDate, err := time.Parse("2006-01-02 15:04:05", p.PlayDate)
			if err != nil {
				slog.Error("Failed to parse play date", "play_uuid", p.UUID, "error", err)
				continue
			}

			err = db.UpsertBoardgamePlay(r.Context(), db.BoardgamePlay{
				ID:     p.UUID,
				GameID: gameUUID,
				Date:   playDate,
			})
			if err != nil {
				slog.Error("Failed to upsert boardgame play", "play_uuid", p.UUID, "error", err)
				continue
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
