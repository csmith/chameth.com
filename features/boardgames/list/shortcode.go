package list

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"chameth.com/chameth.com/features/boardgames"
	"chameth.com/chameth.com/features/shortcodes"
	"tailscale.com/tsnet"
)

const (
	shortcodeVersion = 1
	refreshFrequency = 6 * time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"bglist",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[[]entry], error) {
			return retrieve(ctx, ts.HTTPClient(), args)
		},
		render,
	)
}

type entry struct {
	Name       string `json:"name"`
	Year       int    `json:"year"`
	ImagePath  string `json:"image_path"`
	PlayCount  int    `json:"play_count"`
	LastPlayed string `json:"last_played"`
}

func retrieve(ctx context.Context, client *http.Client, _ []string) (shortcodes.Result[[]entry], error) {
	games, err := boardgames.PlayCounts(ctx, client, "", "")
	if err != nil {
		return shortcodes.Result[[]entry]{}, fmt.Errorf("failed to fetch play counts: %w", err)
	}

	images, err := boardgames.EnsureImages(ctx, client, games)
	if err != nil {
		return shortcodes.Result[[]entry]{}, fmt.Errorf("failed to rehost box art: %w", err)
	}

	entries := make([]entry, 0, len(games))
	for _, g := range games {
		e := entry{
			Name:       g.Name,
			PlayCount:  g.PlayCount,
			LastPlayed: g.LastPlayed,
		}
		if g.BggID != nil {
			e.ImagePath = images[*g.BggID]
		}
		if g.Year != nil {
			e.Year = *g.Year
		}
		entries = append(entries, e)
	}

	return shortcodes.Result[[]entry]{
		Data:      entries,
		RefreshAt: shortcodes.NextRefresh(refreshFrequency, time.Time{}),
	}, nil
}

func render(_ []string, entries []entry, _ *shortcodes.Context) (string, error) {
	games := make([]Game, len(entries))
	for i, e := range entries {
		games[i] = Game{
			Position:   i + 1,
			Name:       e.Name,
			Year:       e.Year,
			ImagePath:  e.ImagePath,
			PlayCount:  e.PlayCount,
			LastPlayed: e.LastPlayed,
		}
	}

	return renderTemplate(Data{Games: games})
}
