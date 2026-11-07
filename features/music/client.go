package music

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"chameth.com/chameth.com/external/twowords"
)

var bonusMetal = twowords.Service{Name: "Bonus Metal", BaseURL: "https://bm.yak-wall.ts.net"}

type NowPlaying struct {
	PlayedAt        time.Time `json:"played_at"`
	Track           string    `json:"track"`
	TrackSubsonicID string    `json:"track_subsonic_id"`
	Album           string    `json:"album"`
	AlbumID         int       `json:"album_id"`
	AlbumSubsonicID string    `json:"album_subsonic_id"`
	Artist          string    `json:"artist"`
	Cover           string    `json:"cover"`
}

// Album is an entry from the insights endpoints.
type Album struct {
	Position         int    `json:"position"`
	Movement         string `json:"movement"`
	PreviousPosition *int   `json:"previous_position"`
	ID               int    `json:"id"`
	SubsonicID       string `json:"subsonic_id"`
	Name             string `json:"name"`
	Artist           string `json:"artist"`
	Year             *int   `json:"year"`
	TrackCount       int    `json:"track_count"`
	PlayCount        int    `json:"play_count"`
	Cover            string `json:"cover"`
}

type Artist struct {
	Position   int    `json:"position"`
	ID         int    `json:"id"`
	SubsonicID string `json:"subsonic_id"`
	Name       string `json:"name"`
	AlbumCount int    `json:"album_count"`
	TrackCount int    `json:"track_count"`
	PlayCount  int    `json:"play_count"`
	Cover      string `json:"cover"`
}

// GetNowPlaying returns the most recent play, or nil when nothing has been
// played yet.
func GetNowPlaying(ctx context.Context, client *http.Client) (*NowPlaying, error) {
	body, _, err := bonusMetal.Get(ctx, client, "/api/nowplaying", nil)
	if errors.Is(err, twowords.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var np NowPlaying
	if err := json.Unmarshal(body, &np); err != nil {
		return nil, fmt.Errorf("failed to decode now playing: %w", err)
	}
	return &np, nil
}

// TopAlbums returns the all-time album ranking; limit of 0 means no limit.
func TopAlbums(ctx context.Context, client *http.Client, limit int) ([]Album, error) {
	body, _, err := bonusMetal.Get(ctx, client, "/api/insights/top-albums", twowords.Query{"limit": limit})
	if err != nil {
		return nil, err
	}

	var res struct {
		Albums []Album `json:"albums"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode top albums: %w", err)
	}
	return res.Albums, nil
}

// TopArtists returns the all-time artist ranking; limit of 0 means no limit.
func TopArtists(ctx context.Context, client *http.Client, limit int) ([]Artist, error) {
	body, _, err := bonusMetal.Get(ctx, client, "/api/insights/top-artists", twowords.Query{"limit": limit})
	if err != nil {
		return nil, err
	}

	var res struct {
		Artists []Artist `json:"artists"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode top artists: %w", err)
	}
	return res.Artists, nil
}

// NewAlbums returns the albums first played in the half-open date range.
func NewAlbums(ctx context.Context, client *http.Client, start, end string) ([]Album, error) {
	body, _, err := bonusMetal.Get(ctx, client, "/api/insights/new-albums", twowords.Query{"start": start, "end": end})
	if err != nil {
		return nil, err
	}

	var res struct {
		Albums []Album `json:"albums"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode new albums: %w", err)
	}
	return res.Albums, nil
}

// AlbumChart returns the most played albums of the half-open date range,
// ranked with movement versus the preceding same-length window.
func AlbumChart(ctx context.Context, client *http.Client, start, end string, limit int) ([]Album, error) {
	body, _, err := bonusMetal.Get(
		ctx,
		client,
		"/api/insights/album-chart",
		twowords.Query{"start": start, "end": end, "limit": limit},
	)
	if err != nil {
		return nil, err
	}

	var res struct {
		Albums []Album `json:"albums"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode album chart: %w", err)
	}
	return res.Albums, nil
}

// FetchImage retrieves artwork; path is the root-relative cover path
// reported by the service.
func FetchImage(ctx context.Context, client *http.Client, path string) ([]byte, string, error) {
	body, header, err := bonusMetal.Get(ctx, client, path, nil)
	if err != nil {
		return nil, "", err
	}
	return body, header.Get("Content-Type"), nil
}
