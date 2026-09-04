package longest

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
	longestRunVersion   = 1
	longestCycleVersion = 1
	refreshFrequency    = 6 * time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"longestrun",
		longestRunVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[*record], error) {
			return retrieveRun(ctx, ts.HTTPClient(), args)
		},
		renderRun,
	)
	mgr.RegisterData(
		"longestcycle",
		longestCycleVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[*record], error) {
			return retrieveCycle(ctx, ts.HTTPClient(), args)
		},
		renderCycle,
	)
}

type record struct {
	Activity  string    `json:"activity"`
	DistanceM float64   `json:"distance_m"`
	StartTime time.Time `json:"start_time"`
}

func retrieveRun(ctx context.Context, client *http.Client, _ []string) (shortcodes.Result[*record], error) {
	longest, err := workouts.GetDistanceRecord(ctx, client, "run", "", "")
	if err != nil {
		return shortcodes.Result[*record]{}, fmt.Errorf("failed to fetch distance record: %w", err)
	}

	refreshAt := shortcodes.NextRefresh(refreshFrequency, time.Time{})
	if longest == nil {
		return shortcodes.Result[*record]{RefreshAt: refreshAt}, nil
	}

	return shortcodes.Result[*record]{
		Data: &record{
			Activity:  longest.Name,
			DistanceM: longest.RankingDistanceM(),
			StartTime: longest.StartTime,
		},
		RefreshAt: refreshAt,
	}, nil
}

func renderRun(_ []string, r *record, _ *shortcodes.Context) (string, error) {
	if r == nil {
		return "", nil
	}
	return renderTemplate(buildData("Longest run", r, runMilestones))
}

func retrieveCycle(ctx context.Context, client *http.Client, _ []string) (shortcodes.Result[*record], error) {
	longest, err := workouts.GetDistanceRecord(ctx, client, "cycle", "", "")
	if err != nil {
		return shortcodes.Result[*record]{}, fmt.Errorf("failed to fetch distance record: %w", err)
	}

	refreshAt := shortcodes.NextRefresh(refreshFrequency, time.Time{})
	if longest == nil {
		return shortcodes.Result[*record]{RefreshAt: refreshAt}, nil
	}

	return shortcodes.Result[*record]{
		Data: &record{
			Activity:  longest.Name,
			DistanceM: longest.DistanceM,
			StartTime: longest.StartTime,
		},
		RefreshAt: refreshAt,
	}, nil
}

func renderCycle(_ []string, r *record, _ *shortcodes.Context) (string, error) {
	if r == nil {
		return "", nil
	}
	return renderTemplate(buildData("Longest bike ride", r, cycleMilestones))
}
