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

// refreshFrequency is how often the PB table is refreshed from the
// Pompei Band API. It shows all-time records, so there is no cutoff.
const refreshFrequency = 6 * time.Hour

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	shortcodes.RegisterData(mgr, "workoutpbs", 1, &dataShortcode{ts: ts})
}

// dataShortcode fetches the PB table from the Pompei Band API, via the
// shortcodes data cache.
type dataShortcode struct {
	ts *tsnet.Server
}

// record is one cached PB: the current all-time record for a single
// segment distance within the shortcode's activity group.
type record struct {
	DistanceM  float64 `json:"distance_m"`
	ElapsedS   float64 `json:"elapsed_s"`
	PaceSPerKm float64 `json:"pace_s_per_km"`
	Date       string  `json:"date"`
}

// Retrieve fetches the current PBs for the shortcode's activity group.
func (s *dataShortcode) Retrieve(ctx context.Context, args []string) (shortcodes.Retrieved[[]record], error) {
	group, _, err := parseActivity(args)
	if err != nil {
		return shortcodes.Retrieved[[]record]{}, err
	}

	pbs, err := workouts.PBs(ctx, s.ts.HTTPClient(), group)
	if err != nil {
		return shortcodes.Retrieved[[]record]{}, fmt.Errorf("failed to fetch PBs: %w", err)
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
	return shortcodes.Retrieved[[]record]{Data: records}, nil
}

// RefreshPolicy refreshes every refreshFrequency.
func (s *dataShortcode) RefreshPolicy([]string) shortcodes.RefreshPolicy {
	return shortcodes.RefreshPolicy{Frequency: refreshFrequency}
}

// Render builds the PB table from the cached records.
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

// activities maps the shortcode's activity argument to the activity_group
// slug used by the API, plus a display label for the box title. Only
// groups that track segment PBs are listed.
var activities = map[string]struct{ group, label string }{
	"running": {"run", "Running"},
	"run":     {"run", "Running"},
	"cycling": {"cycle", "Cycling"},
	"cycle":   {"cycle", "Cycling"},
	"walking": {"walk", "Walking"},
	"walk":    {"walk", "Walking"},
}

// parseActivity validates the shortcode's single activity argument,
// returning its API group slug and display label.
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
