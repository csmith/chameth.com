package calendar

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/shortcodes"
	"tailscale.com/tsnet"
)

const (
	shortcodeVersion = 1
	refreshFrequency = time.Hour
	// Once a year is over, reviews in it can still be edited.
	pastRefreshFrequency = 7 * 24 * time.Hour
	// Watches are sometimes logged a day or two late.
	loggingGrace = 2 * 24 * time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"filmcalendar",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[[]films.WatchDay], error) {
			w, err := parseWindow(args, time.Now())
			if err != nil {
				return shortcodes.Result[[]films.WatchDay]{}, err
			}

			days, err := films.GetWatchDays(
				ctx,
				ts.HTTPClient(),
				w.start.Format(time.DateOnly),
				w.end.Format(time.DateOnly),
			)
			if err != nil {
				return shortcodes.Result[[]films.WatchDay]{}, fmt.Errorf("failed to fetch film watch days: %w", err)
			}

			refreshAt := shortcodes.RefreshIn(refreshFrequency)
			if w.year {
				refreshAt = shortcodes.RefreshUntil(refreshFrequency, w.end.AddDate(0, 0, 1).Add(loggingGrace))
				if refreshAt.IsZero() {
					refreshAt = shortcodes.RefreshIn(pastRefreshFrequency)
				}
			}

			return shortcodes.Result[[]films.WatchDay]{
				Data:      days,
				RefreshAt: refreshAt,
			}, nil
		},
		render,
	)
}

type window struct {
	start, end time.Time
	year       bool
	title      string
}

func parseWindow(args []string, now time.Time) (window, error) {
	switch len(args) {
	case 0:
		today := truncateToDay(now)
		return window{
			start: startOfWeek(today).AddDate(0, 0, -(weeks-1)*7),
			end:   today,
			title: "Films watched",
		}, nil
	case 1:
		year, err := strconv.Atoi(args[0])
		if err != nil {
			return window{}, fmt.Errorf("invalid year: %s (expected YYYY)", args[0])
		}
		return window{
			start: time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC),
			year:  true,
			title: fmt.Sprintf("Films watched in %d", year),
		}, nil
	default:
		return window{}, errors.New("filmcalendar takes 0 arguments or 1 (year)")
	}
}

func render(args []string, days []films.WatchDay, _ *shortcodes.Context) (string, error) {
	now := time.Now()
	w, err := parseWindow(args, now)
	if err != nil {
		return "", err
	}

	return renderTemplate(buildData(w, truncateToDay(now), days))
}
