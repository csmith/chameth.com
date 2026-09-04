package nowplaying

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
	refreshFrequency = time.Minute
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"nowplaying",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[*cached], error) {
			return retrieve(ctx, ts.HTTPClient())
		},
		render,
	)
}

func retrieve(ctx context.Context, client *http.Client) (shortcodes.Result[*cached], error) {
	np, err := music.GetNowPlaying(ctx, client)
	if err != nil {
		return shortcodes.Result[*cached]{}, fmt.Errorf("failed to fetch now playing: %w", err)
	}

	refreshAt := shortcodes.RefreshIn(refreshFrequency)
	if np == nil {
		// Nothing has been played yet; keep polling.
		return shortcodes.Result[*cached]{RefreshAt: refreshAt}, nil
	}

	covers, err := music.EnsureAlbumCovers(ctx, client, []music.Album{{
		ID:         np.AlbumID,
		SubsonicID: np.AlbumSubsonicID,
		Name:       np.Album,
		Cover:      np.Cover,
	}})
	if err != nil {
		return shortcodes.Result[*cached]{}, fmt.Errorf("failed to rehost album cover: %w", err)
	}

	return shortcodes.Result[*cached]{
		Data: &cached{
			ArtistName: np.Artist,
			TrackName:  np.Track,
			AlbumName:  np.Album,
			ImagePath:  covers[np.AlbumSubsonicID],
			PlayedAt:   np.PlayedAt,
		},
		RefreshAt: refreshAt,
	}, nil
}
