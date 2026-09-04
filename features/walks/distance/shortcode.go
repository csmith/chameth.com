package distance

import (
	"context"
	"fmt"
	"net/http"
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
		"walkingdistance",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[float64], error) {
			return retrieve(ctx, ts.HTTPClient(), args)
		},
		render,
	)
}

func retrieve(ctx context.Context, client *http.Client, _ []string) (shortcodes.Result[float64], error) {
	summary, err := workouts.ActivitySummary(ctx, client, "", "", "walk")
	if err != nil {
		return shortcodes.Result[float64]{}, fmt.Errorf("failed to fetch walking summary: %w", err)
	}

	distanceKm := 0.0
	if walk, ok := summary["walk"]; ok {
		distanceKm = walk.DistanceM / 1000
		if walk.WalkDistanceM != nil {
			distanceKm = *walk.WalkDistanceM / 1000
		}
	}

	return shortcodes.Result[float64]{
		Data:      distanceKm,
		RefreshAt: shortcodes.RefreshIn(refreshFrequency),
	}, nil
}
