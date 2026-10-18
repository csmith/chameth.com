package ratingdistribution

import (
	"context"
	"fmt"
	"time"

	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/shortcodes"
	"tailscale.com/tsnet"
)

const (
	shortcodeVersion = 1
	refreshFrequency = time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"filmratingdistribution",
		shortcodeVersion,
		func(ctx context.Context, _ []string) (shortcodes.Result[[]films.RatingCount], error) {
			ratings, err := films.GetRatings(ctx, ts.HTTPClient())
			if err != nil {
				return shortcodes.Result[[]films.RatingCount]{}, fmt.Errorf("failed to fetch film ratings: %w", err)
			}

			return shortcodes.Result[[]films.RatingCount]{
				Data:      ratings,
				RefreshAt: shortcodes.RefreshIn(refreshFrequency),
			}, nil
		},
		render,
	)
}
