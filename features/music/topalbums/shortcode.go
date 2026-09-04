package topalbums

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
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
		"topalbums",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[[]Album], error) {
			return retrieve(ctx, ts.HTTPClient(), args)
		},
		render,
	)
}

func retrieve(ctx context.Context, client *http.Client, args []string) (shortcodes.Result[[]Album], error) {
	limit, err := parseArgs(args)
	if err != nil {
		return shortcodes.Result[[]Album]{}, err
	}

	albums, err := music.TopAlbums(ctx, client, limit)
	if err != nil {
		return shortcodes.Result[[]Album]{}, fmt.Errorf("failed to fetch top albums: %w", err)
	}

	images, err := music.EnsureAlbumCovers(ctx, client, albums)
	if err != nil {
		return shortcodes.Result[[]Album]{}, fmt.Errorf("failed to rehost album art: %w", err)
	}

	items := make([]Album, len(albums))
	for i, a := range albums {
		items[i] = Album{
			Position:   a.Position,
			Name:       a.Name,
			ArtistName: a.Artist,
			TrackCount: a.TrackCount,
			PlayCount:  a.PlayCount,
			ImagePath:  images[a.SubsonicID],
		}
	}

	return shortcodes.Result[[]Album]{
		Data:      items,
		RefreshAt: shortcodes.RefreshIn(refreshFrequency),
	}, nil
}

func parseArgs(args []string) (int, error) {
	if len(args) == 0 {
		return 0, nil
	}

	limit, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, fmt.Errorf("invalid topalbums limit: %s", args[0])
	}
	return limit, nil
}
