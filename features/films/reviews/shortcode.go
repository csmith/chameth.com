package reviews

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
		"filmreviews",
		shortcodeVersion,
		func(ctx context.Context, _ []string) (shortcodes.Result[[]films.ReviewEntry], error) {
			entries, err := films.GetReviews(ctx, ts.HTTPClient(), films.ReviewQuery{})
			if err != nil {
				return shortcodes.Result[[]films.ReviewEntry]{}, fmt.Errorf("failed to fetch film reviews: %w", err)
			}

			return shortcodes.Result[[]films.ReviewEntry]{
				Data:      entries,
				RefreshAt: shortcodes.RefreshIn(refreshFrequency),
			}, nil
		},
		render,
	)
}
