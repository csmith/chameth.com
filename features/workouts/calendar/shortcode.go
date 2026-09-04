package calendar

import (
	"context"
	"fmt"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/workouts"
	"tailscale.com/tsnet"
)

// refreshFrequency is how often the calendar's cached data is refreshed
// from the Pompei Band API.
const refreshFrequency = 6 * time.Hour

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	shortcodes.RegisterData(mgr, "workoutcalendar", 1, &dataShortcode{ts: ts})
}

// dataShortcode fetches the calendar's data from the Pompei Band API,
// via the shortcodes data cache.
type dataShortcode struct {
	ts *tsnet.Server
}

// dayEntry is one day's rolled-up activity data: the activity count and
// the per-group distance totals (already folded into the calendar's
// run/cycle/walk/other buckets) that drive the stripes and tooltips.
type dayEntry struct {
	Date      string             `json:"date"`
	Count     int                `json:"count"`
	Distances map[string]float64 `json:"distances"`
}

// window is the parsed form of the shortcode arguments: the visible date
// span plus the requested range (zero when no range was given).
type window struct {
	start, end           time.Time
	rangeStart, rangeEnd time.Time
	title                string
}

// parseWindow interprets the shortcode arguments: no arguments selects the
// last 16 weeks ending today (UTC); two YYYY-MM-DD arguments select a
// window ending at the end date and extended back to the earlier of the
// start date or 16 weeks, with the requested range highlighted.
func parseWindow(args []string) (window, error) {
	var w window

	switch len(args) {
	case 0:
		now := time.Now().UTC()
		w.end = truncateToDay(now)
		w.start = startOfWeek(now).AddDate(0, 0, -(weeks-1)*7)
		w.title = "Activity calendar"
	case 2:
		var err error
		w.rangeStart, err = time.Parse("2006-01-02", args[0])
		if err != nil {
			return w, fmt.Errorf("invalid start date: %s (expected YYYY-MM-DD)", args[0])
		}
		w.rangeEnd, err = time.Parse("2006-01-02", args[1])
		if err != nil {
			return w, fmt.Errorf("invalid end date: %s (expected YYYY-MM-DD)", args[1])
		}
		w.end = w.rangeEnd
		w.start = w.rangeEnd.AddDate(0, 0, -7*weeks)
		if w.rangeStart.Before(w.start) {
			w.start = w.rangeStart
		}
		w.title = fmt.Sprintf("Activity calendar %s - %s",
			w.rangeStart.Format("2006-01-02"), w.rangeEnd.Format("2006-01-02"))
	default:
		return w, fmt.Errorf("workoutcalendar requires 0 or 2 arguments (start_date, end_date) in YYYY-MM-DD format")
	}

	return w, nil
}

// Retrieve fetches the per-day activity roll-ups for the window from the
// Pompei Band API, folding each day's groups into the calendar's distance
// buckets.
func (s *dataShortcode) Retrieve(ctx context.Context, args []string) (shortcodes.Result[[]dayEntry], error) {
	w, err := parseWindow(args)
	if err != nil {
		return shortcodes.Result[[]dayEntry]{}, err
	}

	days, err := workouts.ActivityDays(ctx, s.ts.HTTPClient(),
		w.start.Format("2006-01-02"), w.end.AddDate(0, 0, 1).Format("2006-01-02"), "")
	if err != nil {
		return shortcodes.Result[[]dayEntry]{}, fmt.Errorf("failed to fetch activity days: %w", err)
	}

	entries := make([]dayEntry, 0, len(days))
	for _, day := range days {
		entry := dayEntry{Date: day.Date, Distances: map[string]float64{}}
		for group, summary := range day.Groups {
			dist := summary.DistanceM
			if group == "run" && summary.RunDistanceM != nil {
				// A run-grouped activity's walking warmup/cooldown
				// intervals shouldn't inflate the run stripe; count only
				// the running distance.
				dist = *summary.RunDistanceM
			}
			entry.Distances[calendarGroup(group)] += dist
			entry.Count += summary.Count
		}
		entries = append(entries, entry)
	}

	var cutoff time.Time
	if !w.rangeEnd.IsZero() {
		cutoff = w.rangeEnd.AddDate(0, 0, 2)
	}

	return shortcodes.Result[[]dayEntry]{
		Data:      entries,
		RefreshAt: shortcodes.NextRefresh(refreshFrequency, cutoff),
	}, nil
}

// Render builds the calendar grid from the cached day entries.
func (s *dataShortcode) Render(args []string, entries []dayEntry, _ *shortcodes.Context) (string, error) {
	w, err := parseWindow(args)
	if err != nil {
		return "", err
	}

	return renderTemplate(buildData(w.title, w.start, w.end, w.rangeStart, w.rangeEnd, entries))
}
