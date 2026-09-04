package achievements

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/wow"
	"tailscale.com/tsnet"
)

// refreshFrequency is how often recent achievements are refreshed from
// the Ogre Stream API.
const refreshFrequency = 4 * time.Hour

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	shortcodes.RegisterData(mgr, "wowachievements", 1, &dataShortcode{ts: ts})
}

// dataShortcode fetches the account's recent achievements from the Ogre
// Stream API, via the shortcodes data cache.
type dataShortcode struct {
	ts *tsnet.Server
}

// parseArgs interprets the shortcode's single optional argument: the
// maximum number of achievements to show (the service clamps it to 1-100).
func parseArgs(args []string) (int, error) {
	if len(args) == 0 {
		return 10, nil
	}

	limit, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, fmt.Errorf("invalid limit %s: %w", args[0], err)
	}
	if limit < 1 {
		return 0, fmt.Errorf("invalid limit %s (must be at least 1)", args[0])
	}
	return limit, nil
}

// Retrieve fetches the account's recent achievements, newest first.
func (s *dataShortcode) Retrieve(ctx context.Context, args []string) (shortcodes.Result[[]Achievement], error) {
	limit, err := parseArgs(args)
	if err != nil {
		return shortcodes.Result[[]Achievement]{}, err
	}

	recent, err := wow.RecentAchievements(ctx, s.ts.HTTPClient(), limit)
	if err != nil {
		return shortcodes.Result[[]Achievement]{}, fmt.Errorf("failed to fetch achievements: %w", err)
	}

	achievements := make([]Achievement, len(recent))
	for i, a := range recent {
		achievements[i] = Achievement{
			ID:          a.ID,
			Name:        a.Name,
			CompletedAt: a.CompletedAt,
		}
	}

	return shortcodes.Result[[]Achievement]{
		Data:      achievements,
		RefreshAt: shortcodes.NextRefresh(refreshFrequency, time.Time{}),
	}, nil
}

// Render builds the achievement list from the cached achievements.
func (s *dataShortcode) Render(_ []string, achievements []Achievement, _ *shortcodes.Context) (string, error) {
	return renderTemplate(Data{Achievements: achievements})
}
