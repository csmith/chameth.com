package watched

import (
	"context"
	"errors"
	"fmt"
	"time"

	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/shortcodes"
	"tailscale.com/tsnet"
)

const (
	shortcodeVersion = 1
	refreshFrequency = time.Hour
	// Once a range is over, reviews in it can still be edited.
	pastRefreshFrequency = 7 * 24 * time.Hour
	// Watches are sometimes logged a day or two late.
	loggingGrace = 2 * 24 * time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"watchedfilms",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[[]films.ReviewEntry], error) {
			start, end, err := parseArgs(args)
			if err != nil {
				return shortcodes.Result[[]films.ReviewEntry]{}, err
			}

			entries, err := films.GetReviews(
				ctx,
				ts.HTTPClient(),
				films.ReviewQuery{
					Start:     start.Format(time.DateOnly),
					End:       end.Format(time.DateOnly),
					Ascending: true,
				},
			)
			if err != nil {
				return shortcodes.Result[[]films.ReviewEntry]{}, fmt.Errorf("failed to fetch watched films: %w", err)
			}

			refreshAt := shortcodes.RefreshUntil(refreshFrequency, end.Add(loggingGrace))
			if refreshAt.IsZero() {
				refreshAt = shortcodes.RefreshIn(pastRefreshFrequency)
			}

			return shortcodes.Result[[]films.ReviewEntry]{
				Data:      entries,
				RefreshAt: refreshAt,
			}, nil
		},
		render,
	)
}

func parseArgs(args []string) (start, end time.Time, err error) {
	if len(args) < 2 {
		return start, end, errors.New("watchedfilms requires 2 arguments (start_date, end_date) in YYYY-MM-DD format")
	}

	start, err = time.Parse(time.DateOnly, args[0])
	if err != nil {
		return start, end, fmt.Errorf("invalid start date: %s (expected YYYY-MM-DD)", args[0])
	}

	end, err = time.Parse(time.DateOnly, args[1])
	if err != nil {
		return start, end, fmt.Errorf("invalid end date: %s (expected YYYY-MM-DD)", args[1])
	}
	return start, end, nil
}
