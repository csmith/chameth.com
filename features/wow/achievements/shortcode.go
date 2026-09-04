package achievements

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/wow"
	"tailscale.com/tsnet"
)

const (
	shortcodeVersion = 1
	refreshFrequency = 4 * time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"wowachievements",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[[]Achievement], error) {
			return retrieve(ctx, ts.HTTPClient(), args)
		},
		render,
	)
}

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

func retrieve(ctx context.Context, client *http.Client, args []string) (shortcodes.Result[[]Achievement], error) {
	limit, err := parseArgs(args)
	if err != nil {
		return shortcodes.Result[[]Achievement]{}, err
	}

	recent, err := wow.RecentAchievements(ctx, client, limit)
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
		RefreshAt: shortcodes.RefreshIn(refreshFrequency),
	}, nil
}

func render(_ []string, achievements []Achievement, _ *shortcodes.Context) (string, error) {
	return renderTemplate(Data{Achievements: achievements})
}
