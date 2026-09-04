package played

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
		"playedbgs",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[[]entry], error) {
			return retrieve(ctx, ts.HTTPClient(), args)
		},
		render,
	)
}

type entry struct {
	Name      string `json:"name"`
	Year      int    `json:"year"`
	ImagePath string `json:"image_path"`
	PlayCount int    `json:"play_count"`
}

func parseArgs(args []string) (start, end time.Time, err error) {
	if len(args) < 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("playedbgs requires 2 arguments (start_date, end_date) in YYYY-MM-DD format")
	}

	start, err = time.Parse("2006-01-02", args[0])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %s (expected YYYY-MM-DD)", args[0])
	}

	end, err = time.Parse("2006-01-02", args[1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %s (expected YYYY-MM-DD)", args[1])
	}

	return start, end, nil
}

func retrieve(ctx context.Context, client *http.Client, args []string) (shortcodes.Result[[]entry], error) {
	startDate, endDate, err := parseArgs(args)
	if err != nil {
		return shortcodes.Result[[]entry]{}, err
	}

	games, err := boardgames.PlayCounts(ctx, client,
		startDate.Format("2006-01-02"), endDate.AddDate(0, 0, 1).Format("2006-01-02"))
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
			Name:      g.Name,
			PlayCount: g.PlayCount,
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
		RefreshAt: shortcodes.RefreshUntil(refreshFrequency, endDate.AddDate(0, 0, 2)),
	}, nil
}

func render(_ []string, entries []entry, _ *shortcodes.Context) (string, error) {
	games := make([]Game, len(entries))
	for i, e := range entries {
		games[i] = Game(e)
	}

	return renderTemplate(Data{Games: games})
}
