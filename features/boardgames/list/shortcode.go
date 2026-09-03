package list

import (
	"context"
	"fmt"
	"time"

	"chameth.com/chameth.com/external/magicmeters"
	"chameth.com/chameth.com/features/boardgames"
	"chameth.com/chameth.com/features/shortcodes"
	"tailscale.com/tsnet"
)

// refreshFrequency is how often the list is refreshed from the Magic
// Meters API. It covers all time, so there is no cutoff.
const refreshFrequency = 6 * time.Hour

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	shortcodes.RegisterData(mgr, "bglist", 1, &dataShortcode{ts: ts})
}

// dataShortcode fetches the play-count list from the Magic Meters API,
// via the shortcodes data cache.
type dataShortcode struct {
	ts *tsnet.Server
}

// entry is one cached game: its all-time play count, when it was last
// played, and the local path of its rehosted box art (empty when the
// game has none).
type entry struct {
	Name       string `json:"name"`
	Year       int    `json:"year"`
	ImagePath  string `json:"image_path"`
	PlayCount  int    `json:"play_count"`
	LastPlayed string `json:"last_played"`
}

// Retrieve fetches the all-time play counts. The API orders games by play
// count descending, then name, which becomes the list's position
// numbering.
func (s *dataShortcode) Retrieve(ctx context.Context, _ []string) (shortcodes.Retrieved[[]entry], error) {
	client := magicmeters.NewClient(s.ts.HTTPClient())

	games, err := client.PlayCounts(ctx, "", "")
	if err != nil {
		return shortcodes.Retrieved[[]entry]{}, fmt.Errorf("failed to fetch play counts: %w", err)
	}

	images, err := boardgames.EnsureImages(ctx, client, games)
	if err != nil {
		return shortcodes.Retrieved[[]entry]{}, fmt.Errorf("failed to rehost box art: %w", err)
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

	return shortcodes.Retrieved[[]entry]{Data: entries}, nil
}

// RefreshPolicy refreshes every refreshFrequency.
func (s *dataShortcode) RefreshPolicy([]string) shortcodes.RefreshPolicy {
	return shortcodes.RefreshPolicy{Frequency: refreshFrequency}
}

// Render builds the positioned list from the cached entries.
func (s *dataShortcode) Render(_ []string, entries []entry, _ *shortcodes.Context) (string, error) {
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
