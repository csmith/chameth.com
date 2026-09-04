package newalbums

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"chameth.com/chameth.com/features/music"
	"chameth.com/chameth.com/features/shortcodes"
	"tailscale.com/tsnet"
)

const (
	shortcodeVersion = 1
	refreshFrequency = 4 * time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"newalbums",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[[]Album], error) {
			return retrieve(ctx, ts.HTTPClient(), args)
		},
		render,
	)
}

func retrieve(ctx context.Context, client *http.Client, args []string) (shortcodes.Result[[]Album], error) {
	start, end, err := parseArgs(args)
	if err != nil {
		return shortcodes.Result[[]Album]{}, err
	}

	// The service takes a half-open range; the shortcode's end date is inclusive.
	albums, err := music.NewAlbums(ctx, client,
		start.Format("2006-01-02"), end.AddDate(0, 0, 1).Format("2006-01-02"))
	if err != nil {
		return shortcodes.Result[[]Album]{}, fmt.Errorf("failed to fetch new albums: %w", err)
	}

	images, err := music.EnsureAlbumCovers(ctx, client, albums)
	if err != nil {
		return shortcodes.Result[[]Album]{}, fmt.Errorf("failed to rehost album art: %w", err)
	}

	items := make([]Album, len(albums))
	for i, a := range albums {
		items[i] = Album{
			Name:       a.Name,
			ArtistName: a.Artist,
			ImagePath:  images[a.SubsonicID],
		}
	}

	return shortcodes.Result[[]Album]{
		Data:      items,
		RefreshAt: shortcodes.RefreshUntil(refreshFrequency, end.AddDate(0, 0, 2)),
	}, nil
}

func parseArgs(args []string) (start, end time.Time, err error) {
	if len(args) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("newalbums requires 2 arguments (start date, end date)")
	}

	start, err = time.Parse("2006-01-02", args[0])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %s (expected YYYY-MM-DD)", args[0])
	}
	end, err = time.Parse("2006-01-02", args[1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %s (expected YYYY-MM-DD)", args[1])
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("end date %s is before start date %s", args[1], args[0])
	}
	return start, end, nil
}
