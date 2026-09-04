package calendar

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
	refreshFrequency = 6 * time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"workoutcalendar",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[[]dayEntry], error) {
			return retrieve(ctx, ts.HTTPClient(), args)
		},
		render,
	)
}

type dayEntry struct {
	Date      string             `json:"date"`
	Count     int                `json:"count"`
	Distances map[string]float64 `json:"distances"`
}

type window struct {
	start, end           time.Time
	rangeStart, rangeEnd time.Time
	title                string
}

// A named range is extended backwards to show at least 16 weeks.
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

func retrieve(ctx context.Context, client *http.Client, args []string) (shortcodes.Result[[]dayEntry], error) {
	w, err := parseWindow(args)
	if err != nil {
		return shortcodes.Result[[]dayEntry]{}, err
	}

	days, err := workouts.ActivityDays(ctx, client,
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
		RefreshAt: shortcodes.RefreshUntil(refreshFrequency, cutoff),
	}, nil
}

func render(args []string, entries []dayEntry, _ *shortcodes.Context) (string, error) {
	w, err := parseWindow(args)
	if err != nil {
		return "", err
	}

	return renderTemplate(buildData(w.title, w.start, w.end, w.rangeStart, w.rangeEnd, entries))
}
