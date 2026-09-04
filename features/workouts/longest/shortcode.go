package longest

import (
	"context"
	"fmt"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/workouts"
	"tailscale.com/tsnet"
)

// refreshFrequency is how often the records are refreshed from the
// Pompei Band API. They cover all time, so there is no cutoff.
const refreshFrequency = 6 * time.Hour

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	shortcodes.RegisterData(mgr, "longestrun", 1, &runShortcode{ts: ts})
	shortcodes.RegisterData(mgr, "longestcycle", 1, &cycleShortcode{ts: ts})
}

// record is the cached data behind either record: the activity, its
// distance and when it started. A nil record means no qualifying
// activity exists, in which case the shortcode renders nothing.
type record struct {
	Activity  string    `json:"activity"`
	DistanceM float64   `json:"distance_m"`
	StartTime time.Time `json:"start_time"`
}

// runShortcode renders the longestrun shortcode: the longest run ever
// recorded.
type runShortcode struct {
	ts *tsnet.Server
}

// Retrieve fetches the all-time run record. Run records rank on the
// gait-classified running distance (falling back to the recorded total
// when an activity carries no split), so walking breaks inside a run
// don't win the record.
func (s *runShortcode) Retrieve(ctx context.Context, _ []string) (shortcodes.Result[*record], error) {
	longest, err := workouts.GetDistanceRecord(ctx, s.ts.HTTPClient(), "run", "", "")
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

func (s *runShortcode) Render(_ []string, r *record, _ *shortcodes.Context) (string, error) {
	if r == nil {
		return "", nil
	}
	return renderTemplate(buildData("Longest run", r, runMilestones))
}

// cycleShortcode renders the longestcycle shortcode: the longest bike
// ride ever recorded.
type cycleShortcode struct {
	ts *tsnet.Server
}

func (s *cycleShortcode) Retrieve(ctx context.Context, _ []string) (shortcodes.Result[*record], error) {
	longest, err := workouts.GetDistanceRecord(ctx, s.ts.HTTPClient(), "cycle", "", "")
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

func (s *cycleShortcode) Render(_ []string, r *record, _ *shortcodes.Context) (string, error) {
	if r == nil {
		return "", nil
	}
	return renderTemplate(buildData("Longest bike ride", r, cycleMilestones))
}
