package summary

import (
	"context"
	"fmt"
	"sort"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/workouts"
	"tailscale.com/tsnet"
)

// refreshFrequency is how often the summary's cached data is refreshed
// from the Pompei Band API.
const refreshFrequency = 6 * time.Hour

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	shortcodes.RegisterData(mgr, "workoutsummary", 1, &dataShortcode{ts: ts})
}

// dataShortcode fetches the summary's data from the Pompei Band API, via
// the shortcodes data cache.
type dataShortcode struct {
	ts *tsnet.Server
}

// record is the longest single activity (or running interval) within the
// period.
type record struct {
	DistanceM float64   `json:"distance_m"`
	StartTime time.Time `json:"start_time"`
}

// pb is the fastest record-breaking effort for one segment distance
// within the period, along with the record it displaced.
type pb struct {
	Group            string   `json:"group"`
	DistanceM        float64  `json:"distance_m"`
	ElapsedS         float64  `json:"elapsed_s"`
	PreviousElapsedS *float64 `json:"previous_elapsed_s"`
}

// data is everything the summary renders, derived from the API's
// activity-summary, distance-records, activities and pb-events endpoints.
type data struct {
	CycleCount     int     `json:"cycle_count"`
	CycleDistanceM float64 `json:"cycle_distance_m"`
	CycleDurationS float64 `json:"cycle_duration_s"`
	RunCount       int     `json:"run_count"`
	RunDistanceM   float64 `json:"run_distance_m"`
	RunDurationS   float64 `json:"run_duration_s"`
	LongestCycle   *record `json:"longest_cycle"`
	LongestRun     *record `json:"longest_run"`
	PBs            []pb    `json:"pbs"`
}

// parseArgs interprets the shortcode's two YYYY-MM-DD arguments. The
// window sent to the API is half-open: endExclusive extends the range
// through the whole of the end date.
func parseArgs(args []string) (start, end time.Time, err error) {
	if len(args) < 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("workoutsummary requires 2 arguments (start_date, end_date) in YYYY-MM-DD format")
	}

	start, err = time.Parse("2006-01-02", args[0])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %s (expected YYYY-MM-DD)", args[0])
	}

	end, err = time.Parse("2006-01-02", args[1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %s (expected YYYY-MM-DD)", args[1])
	}

	return start, end, nil
}

// Retrieve assembles the period's data from four API calls: per-group
// totals, the longest ride, the longest running interval and the record-
// breaking efforts.
func (s *dataShortcode) Retrieve(ctx context.Context, args []string) (shortcodes.Retrieved[data], error) {
	startDate, endDate, err := parseArgs(args)
	if err != nil {
		return shortcodes.Retrieved[data]{}, err
	}

	start := startDate.Format("2006-01-02")
	endExclusive := endDate.AddDate(0, 0, 1).Format("2006-01-02")
	client := s.ts.HTTPClient()

	var d data

	groups, err := workouts.ActivitySummary(ctx, client, start, endExclusive, "")
	if err != nil {
		return shortcodes.Retrieved[data]{}, fmt.Errorf("failed to fetch activity summary: %w", err)
	}
	if g, ok := groups["cycle"]; ok {
		d.CycleCount = g.Count
		d.CycleDistanceM = g.DistanceM
		d.CycleDurationS = g.DurationS
	}
	if g, ok := groups["run"]; ok {
		d.RunCount = g.Count
		d.RunDurationS = g.DurationS
		// The running total is the run group's gait-classified running
		// distance. Deliberate running efforts land in the run group, so
		// other groups' incidental run splits (GPS drift, a jog across a
		// road) don't count towards it; fall back to the group's recorded
		// total when no activity carries a split.
		d.RunDistanceM = g.DistanceM
		if g.RunDistanceM != nil {
			d.RunDistanceM = *g.RunDistanceM
		}
	}

	longestCycle, err := workouts.GetDistanceRecord(ctx, client, "cycle", start, endExclusive)
	if err != nil {
		return shortcodes.Retrieved[data]{}, fmt.Errorf("failed to fetch longest ride: %w", err)
	}
	if longestCycle != nil {
		d.LongestCycle = &record{DistanceM: longestCycle.DistanceM, StartTime: longestCycle.StartTime}
	}

	longestRun, err := workouts.GetDistanceRecord(ctx, client, "run", start, endExclusive)
	if err != nil {
		return shortcodes.Retrieved[data]{}, fmt.Errorf("failed to fetch longest run: %w", err)
	}
	if longestRun != nil {
		// Run records rank on the gait-classified running distance
		// (falling back to the recorded total), matching how the period's
		// longest run has always been measured.
		d.LongestRun = &record{DistanceM: longestRun.RankingDistanceM(), StartTime: longestRun.StartTime}
	}

	events, err := workouts.PBEvents(ctx, client, start, endExclusive)
	if err != nil {
		return shortcodes.Retrieved[data]{}, fmt.Errorf("failed to fetch PB events: %w", err)
	}
	d.PBs = fastestPBs(events)

	return shortcodes.Retrieved[data]{Data: d}, nil
}

// fastestPBs condenses the window's record-breaking efforts into one entry
// per (group, distance): the fastest time achieved, with the previous
// record taken from the earliest improvement in the window (so it reflects
// the record that stood before the period began, even if the record was
// beaten several times since). Events arrive chronological; only the
// cycle and run groups are reported, as before.
func fastestPBs(events []workouts.PBEvent) []pb {
	byKey := map[string]map[float64]*pb{}
	for _, e := range events {
		if e.Group != "cycle" && e.Group != "run" {
			continue
		}
		group, ok := byKey[e.Group]
		if !ok {
			group = map[float64]*pb{}
			byKey[e.Group] = group
		}
		if existing, ok := group[e.DistanceM]; ok {
			if e.ElapsedS < existing.ElapsedS {
				existing.ElapsedS = e.ElapsedS
			}
		} else {
			group[e.DistanceM] = &pb{
				Group:            e.Group,
				DistanceM:        e.DistanceM,
				ElapsedS:         e.ElapsedS,
				PreviousElapsedS: e.PreviousElapsedS,
			}
		}
	}

	var result []pb
	for _, distances := range byKey {
		for _, p := range distances {
			result = append(result, *p)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Group != result[j].Group {
			return result[i].Group < result[j].Group
		}
		return result[i].DistanceM < result[j].DistanceM
	})
	return result
}

// RefreshPolicy refreshes every refreshFrequency, stopping one day past
// the end of the period (the end date plus the whole following day), so
// late-reported activities still show up before the data freezes.
func (s *dataShortcode) RefreshPolicy(args []string) shortcodes.RefreshPolicy {
	policy := shortcodes.RefreshPolicy{Frequency: refreshFrequency}
	if _, end, err := parseArgs(args); err == nil {
		policy.Cutoff = end.AddDate(0, 0, 2)
	}
	return policy
}

// Render builds the summary sections from the cached data.
func (s *dataShortcode) Render(args []string, d data, _ *shortcodes.Context) (string, error) {
	startDate, endDate, err := parseArgs(args)
	if err != nil {
		return "", err
	}

	dateRange := fmt.Sprintf("%s – %s", startDate.Format("2 Jan"), endDate.Format("2 Jan 2006"))
	return renderTemplate(buildData(dateRange, d))
}
