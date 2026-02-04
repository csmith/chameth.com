package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"chameth.com/chameth.com/cmd/serve/db"
)

type WorkoutData struct {
	Data struct {
		Workouts []Workout `json:"workouts"`
	} `json:"data"`
}

type Workout struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Start       string   `json:"start"`
	End         string   `json:"end"`
	Duration    float64  `json:"duration"`
	Distance    Quantity `json:"distance"`
	ElevationUp Quantity `json:"elevationUp"`
}

type Quantity struct {
	Qty   float64 `json:"qty"`
	Units string  `json:"units"`
}

func ImportWalksHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var data WorkoutData
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			slog.Error("Failed to decode workout data", "error", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		for _, workout := range data.Data.Workouts {
			if workout.Name != "Outdoor Walk" {
				continue
			}

			startDate, err := time.Parse("2006-01-02 15:04:05 -0700", workout.Start)
			if err != nil {
				slog.Error("Failed to parse start date", "workout_id", workout.ID, "error", err)
				continue
			}

			endDate, err := time.Parse("2006-01-02 15:04:05 -0700", workout.End)
			if err != nil {
				slog.Error("Failed to parse end date", "workout_id", workout.ID, "error", err)
				continue
			}

			err = db.UpsertWalk(r.Context(), db.Walk{
				ExternalID:          workout.ID,
				StartDate:           startDate,
				EndDate:             endDate,
				DurationSeconds:     workout.Duration,
				DistanceKm:          workout.Distance.Qty,
				ElevationGainMeters: workout.ElevationUp.Qty,
			})
			if err != nil {
				slog.Error("Failed to upsert walk", "workout_id", workout.ID, "error", err)
				continue
			}

			slog.Info("Imported walk", "external_id", workout.ID, "start", startDate)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
