package list

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/workouts"
	"tailscale.com/tsnet"
)

const (
	shortcodeVersion = 1
	refreshFrequency = 24 * time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"walks",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[[]walk], error) {
			return retrieve(ctx, ts.HTTPClient(), args)
		},
		render,
	)
}

type walk struct {
	Date       time.Time `json:"date"`
	DurationS  float64   `json:"duration_s"`
	DistanceKm float64   `json:"distance_km"`
	ElevationM float64   `json:"elevation_m"`
}

func retrieve(ctx context.Context, client *http.Client, _ []string) (shortcodes.Result[[]walk], error) {
	activities, err := workouts.Activities(ctx, client, "walk")
	if err != nil {
		return shortcodes.Result[[]walk]{}, fmt.Errorf("failed to fetch walks: %w", err)
	}

	walks := make([]walk, 0, len(activities))
	for _, a := range activities {
		var elevationM float64
		if a.Elevation != nil {
			elevationM = a.Elevation.GainM
		}
		walks = append(walks, walk{
			Date:       a.StartTime,
			DurationS:  a.DurationS,
			DistanceKm: a.WalkingDistanceM() / 1000,
			ElevationM: elevationM,
		})
	}

	// The API returns activities oldest first; the page lists newest first.
	sort.Slice(walks, func(i, j int) bool { return walks[i].Date.After(walks[j].Date) })

	return shortcodes.Result[[]walk]{
		Data:      walks,
		RefreshAt: shortcodes.RefreshIn(refreshFrequency),
	}, nil
}
