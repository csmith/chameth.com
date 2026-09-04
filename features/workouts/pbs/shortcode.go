package pbs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/workouts"
	"tailscale.com/tsnet"
)

const refreshFrequency = 6 * time.Hour

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	shortcodes.RegisterData(mgr, "workoutpbs", 1, &dataShortcode{ts: ts})
}

type dataShortcode struct {
	ts *tsnet.Server
}

type record struct {
	DistanceM  float64 `json:"distance_m"`
	ElapsedS   float64 `json:"elapsed_s"`
	PaceSPerKm float64 `json:"pace_s_per_km"`
	Date       string  `json:"date"`
}

func (s *dataShortcode) Retrieve(ctx context.Context, args []string) (shortcodes.Result[[]record], error) {
	group, _, err := parseActivity(args)
	if err != nil {
		return shortcodes.Result[[]record]{}, err
	}

	pbs, err := workouts.PBs(ctx, s.ts.HTTPClient(), group)
	if err != nil {
		return shortcodes.Result[[]record]{}, fmt.Errorf("failed to fetch PBs: %w", err)
	}

	records := make([]record, 0, len(pbs))
	for _, pb := range pbs {
		records = append(records, record{
			DistanceM:  pb.DistanceM,
			ElapsedS:   pb.ElapsedS,
			PaceSPerKm: pb.PaceSPerKm,
			Date:       pb.Date,
		})
	}
	return shortcodes.Result[[]record]{
		Data:      records,
		RefreshAt: shortcodes.NextRefresh(refreshFrequency, time.Time{}),
	}, nil
}

func (s *dataShortcode) Render(args []string, records []record, _ *shortcodes.Context) (string, error) {
	_, label, err := parseActivity(args)
	if err != nil {
		return "", err
	}
	if len(records) == 0 {
		return "", nil
	}
	return renderTemplate(buildData(label+" PBs", records))
}

var activities = map[string]struct{ group, label string }{
	"running": {"run", "Running"},
	"run":     {"run", "Running"},
	"cycling": {"cycle", "Cycling"},
	"cycle":   {"cycle", "Cycling"},
	"walking": {"walk", "Walking"},
	"walk":    {"walk", "Walking"},
}

func parseActivity(args []string) (group, label string, err error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("workoutpbs requires 1 argument (activity), e.g. \"running\"")
	}
	a, ok := activities[strings.ToLower(args[0])]
	if !ok {
		return "", "", fmt.Errorf("unknown activity: %s", args[0])
	}
	return a.group, a.label, nil
}
