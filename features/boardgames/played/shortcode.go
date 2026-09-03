package played

import (
	"context"
	"fmt"
	"time"

	"chameth.com/chameth.com/external/magicmeters"
	"chameth.com/chameth.com/features/boardgames"
	"chameth.com/chameth.com/features/shortcodes"
	"tailscale.com/tsnet"
)

// refreshFrequency is how often a window's play counts are refreshed from
// the Magic Meters API.
const refreshFrequency = 6 * time.Hour

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	shortcodes.RegisterData(mgr, "playedbgs", 1, &dataShortcode{ts: ts})
}

// dataShortcode fetches a window's play counts from the Magic Meters API,
// via the shortcodes data cache.
type dataShortcode struct {
	ts *tsnet.Server
}

// entry is one cached game's play count within the window, plus the local
// path of its rehosted box art (empty when the game has none).
type entry struct {
	Name      string `json:"name"`
	Year      int    `json:"year"`
	ImagePath string `json:"image_path"`
	PlayCount int    `json:"play_count"`
}

// parseArgs interprets the shortcode's two YYYY-MM-DD arguments. The
// window sent to the API is half-open: endExclusive extends the range
// through the whole of the end date.
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

// Retrieve fetches the window's play counts. The API orders games by play
// count descending, then name.
func (s *dataShortcode) Retrieve(ctx context.Context, args []string) (shortcodes.Retrieved[[]entry], error) {
	startDate, endDate, err := parseArgs(args)
	if err != nil {
		return shortcodes.Retrieved[[]entry]{}, err
	}

	client := magicmeters.NewClient(s.ts.HTTPClient())

	games, err := client.PlayCounts(ctx,
		startDate.Format("2006-01-02"), endDate.AddDate(0, 0, 1).Format("2006-01-02"))
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

	return shortcodes.Retrieved[[]entry]{Data: entries}, nil
}

// RefreshPolicy refreshes every refreshFrequency, stopping one day past
// the end of the window (the end date plus the whole following day), so
// late-logged plays still show up before the data freezes.
func (s *dataShortcode) RefreshPolicy(args []string) shortcodes.RefreshPolicy {
	policy := shortcodes.RefreshPolicy{Frequency: refreshFrequency}
	if _, end, err := parseArgs(args); err == nil {
		policy.Cutoff = end.AddDate(0, 0, 2)
	}
	return policy
}

// Render builds the grid from the cached entries.
func (s *dataShortcode) Render(_ []string, entries []entry, _ *shortcodes.Context) (string, error) {
	games := make([]Game, len(entries))
	for i, e := range entries {
		games[i] = Game(e)
	}

	return renderTemplate(Data{Games: games})
}
