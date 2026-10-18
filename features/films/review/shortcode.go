package review

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
	refreshFrequency = 7 * 24 * time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"filmreview",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[*films.ReviewEntry], error) {
			id, err := parseArgs(args)
			if err != nil {
				return shortcodes.Result[*films.ReviewEntry]{}, err
			}

			// A missing review is cached too, so that it is only rechecked
			// on the normal schedule.
			entry, err := films.GetReview(ctx, ts.HTTPClient(), id)
			if err != nil {
				return shortcodes.Result[*films.ReviewEntry]{}, fmt.Errorf("failed to fetch film review: %w", err)
			}

			return shortcodes.Result[*films.ReviewEntry]{
				Data:      entry,
				RefreshAt: shortcodes.RefreshIn(refreshFrequency),
			}, nil
		},
		render,
	)
}

func parseArgs(args []string) (int, error) {
	if len(args) < 1 {
		return 0, errors.New("filmreview requires at least 1 argument (id)")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, fmt.Errorf("invalid film review ID: %s", args[0])
	}
	return id, nil
}
