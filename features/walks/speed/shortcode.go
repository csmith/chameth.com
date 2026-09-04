package speed

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
		"walkingspeed",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[[]monthSpeed], error) {
			return retrieve(ctx, ts.HTTPClient(), args)
		},
		render,
	)
}

type monthSpeed struct {
	Month    time.Time `json:"month"`
	SpeedKmh float64   `json:"speed_kmh"`
}

func retrieve(ctx context.Context, client *http.Client, _ []string) (shortcodes.Result[[]monthSpeed], error) {
	months, err := workouts.ActivityMonths(ctx, client, "walk")
	if err != nil {
		return shortcodes.Result[[]monthSpeed]{}, fmt.Errorf("failed to fetch walking months: %w", err)
	}

	speeds := make([]monthSpeed, 0, len(months))
	for _, m := range months {
		if m.MaxAverageSpeedMps == nil {
			continue
		}

		month, err := time.Parse("2006-01", m.Month)
		if err != nil {
			return shortcodes.Result[[]monthSpeed]{}, fmt.Errorf("invalid month %q: %w", m.Month, err)
		}

		speeds = append(speeds, monthSpeed{
			Month:    month,
			SpeedKmh: *m.MaxAverageSpeedMps * 3.6,
		})
	}

	// The chart plots by slice index, so enforce chronological order
	// rather than trusting the API's oldest-first ordering.
	sort.Slice(speeds, func(i, j int) bool { return speeds[i].Month.Before(speeds[j].Month) })

	return shortcodes.Result[[]monthSpeed]{
		Data:      speeds,
		RefreshAt: shortcodes.RefreshIn(refreshFrequency),
	}, nil
}
