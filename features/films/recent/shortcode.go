package recent

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
	shortcodeName    = "recentfilms"
	shortcodeVersion = 1
	refreshFrequency = time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		shortcodeName,
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[[]films.ReviewEntry], error) {
			count, err := parseArgs(args)
			if err != nil {
				return shortcodes.Result[[]films.ReviewEntry]{}, err
			}

			entries, err := films.GetReviews(ctx, ts.HTTPClient(), films.ReviewQuery{Limit: count})
			if err != nil {
				return shortcodes.Result[[]films.ReviewEntry]{}, fmt.Errorf("failed to fetch recent film reviews: %w", err)
			}

			return shortcodes.Result[[]films.ReviewEntry]{
				Data:      entries,
				RefreshAt: shortcodes.RefreshIn(refreshFrequency),
			}, nil
		},
		render,
	)
}

// Reviews returns the latest count published reviews, sharing the cache
// used by the recentfilms shortcode.
func Reviews(ctx context.Context, mgr *shortcodes.Manager, count int) ([]films.ReviewEntry, error) {
	return shortcodes.CachedData[[]films.ReviewEntry](mgr, ctx, shortcodeName, []string{strconv.Itoa(count)})
}

func parseArgs(args []string) (int, error) {
	if len(args) < 1 {
		return 0, errors.New("recentfilms requires at least 1 argument (count)")
	}

	count, err := strconv.Atoi(args[0])
	if err != nil || count < 1 {
		return 0, fmt.Errorf("invalid recent films count: %s", args[0])
	}
	return count, nil
}
