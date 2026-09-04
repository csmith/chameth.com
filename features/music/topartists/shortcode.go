package topartists

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
		"topartists",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[[]Artist], error) {
			return retrieve(ctx, ts.HTTPClient(), args)
		},
		render,
	)
}

func retrieve(ctx context.Context, client *http.Client, args []string) (shortcodes.Result[[]Artist], error) {
	limit, err := parseArgs(args)
	if err != nil {
		return shortcodes.Result[[]Artist]{}, err
	}

	artists, err := music.TopArtists(ctx, client, limit)
	if err != nil {
		return shortcodes.Result[[]Artist]{}, fmt.Errorf("failed to fetch top artists: %w", err)
	}

	images, err := music.EnsureArtistCovers(ctx, client, artists)
	if err != nil {
		return shortcodes.Result[[]Artist]{}, fmt.Errorf("failed to rehost artist images: %w", err)
	}

	items := make([]Artist, len(artists))
	for i, a := range artists {
		items[i] = Artist{
			Position:   a.Position,
			Name:       a.Name,
			TrackCount: a.TrackCount,
			AlbumCount: a.AlbumCount,
			PlayCount:  a.PlayCount,
			ImagePath:  images[a.SubsonicID],
		}
	}

	return shortcodes.Result[[]Artist]{
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
		return 0, fmt.Errorf("invalid topartists limit: %s", args[0])
	}
	return limit, nil
}
