package summary

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
	refreshFrequency = 6 * time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"workoutsummary",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[data], error) {
			return retrieve(ctx, ts.HTTPClient(), args)
		},
		render,
	)
}

type record struct {
	DistanceM float64   `json:"distance_m"`
	StartTime time.Time `json:"start_time"`
}

type pb struct {
	Group            string   `json:"group"`
	DistanceM        float64  `json:"distance_m"`
	ElapsedS         float64  `json:"elapsed_s"`
	PreviousElapsedS *float64 `json:"previous_elapsed_s"`
}

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

func retrieve(ctx context.Context, client *http.Client, args []string) (shortcodes.Result[data], error) {
	startDate, endDate, err := parseArgs(args)
	if err != nil {
		return shortcodes.Result[data]{}, err
	}

	start := startDate.Format("2006-01-02")
	endExclusive := endDate.AddDate(0, 0, 1).Format("2006-01-02")
	var d data

	groups, err := workouts.ActivitySummary(ctx, client, start, endExclusive, "")
	if err != nil {
		return shortcodes.Result[data]{}, fmt.Errorf("failed to fetch activity summary: %w", err)
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
		return shortcodes.Result[data]{}, fmt.Errorf("failed to fetch longest ride: %w", err)
	}
	if longestCycle != nil {
		d.LongestCycle = &record{DistanceM: longestCycle.DistanceM, StartTime: longestCycle.StartTime}
	}

	longestRun, err := workouts.GetDistanceRecord(ctx, client, "run", start, endExclusive)
	if err != nil {
		return shortcodes.Result[data]{}, fmt.Errorf("failed to fetch longest run: %w", err)
	}
	if longestRun != nil {
		d.LongestRun = &record{DistanceM: longestRun.RankingDistanceM(), StartTime: longestRun.StartTime}
	}

	events, err := workouts.PBEvents(ctx, client, start, endExclusive)
	if err != nil {
		return shortcodes.Result[data]{}, fmt.Errorf("failed to fetch PB events: %w", err)
	}
	d.PBs = fastestPBs(events)

	return shortcodes.Result[data]{
		Data:      d,
		RefreshAt: shortcodes.NextRefresh(refreshFrequency, endDate.AddDate(0, 0, 2)),
	}, nil
}

// Keep the fastest result per distance while retaining the record that
// stood before the window's first improvement.
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

func render(args []string, d data, _ *shortcodes.Context) (string, error) {
	startDate, endDate, err := parseArgs(args)
	if err != nil {
		return "", err
	}

	dateRange := fmt.Sprintf("%s – %s", startDate.Format("2 Jan"), endDate.Format("2 Jan 2006"))
	return renderTemplate(buildData(dateRange, d))
}
