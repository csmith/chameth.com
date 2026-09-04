package music

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const bonusMetalBaseURL = "https://bm.yak-wall.ts.net"

var errNotFound = errors.New("not found")

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

// Album is an entry from the insights endpoints. The chart fields are only
// populated by the album chart; ID and SubsonicID key the rehosted cover art.
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
	body, _, err := bonusMetalGet(ctx, client, "/api/nowplaying", nil)
	if errors.Is(err, errNotFound) {
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
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}

	body, _, err := bonusMetalGet(ctx, client, "/api/insights/top-albums", query)
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
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}

	body, _, err := bonusMetalGet(ctx, client, "/api/insights/top-artists", query)
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
	query := url.Values{"start": {start}, "end": {end}}

	body, _, err := bonusMetalGet(ctx, client, "/api/insights/new-albums", query)
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
	query := url.Values{"start": {start}, "end": {end}}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}

	body, _, err := bonusMetalGet(ctx, client, "/api/insights/album-chart", query)
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
	body, header, err := bonusMetalGet(ctx, client, path, nil)
	if err != nil {
		return nil, "", err
	}
	return body, header.Get("Content-Type"), nil
}

func bonusMetalGet(ctx context.Context, client *http.Client, path string, query url.Values) ([]byte, http.Header, error) {
	u := strings.TrimRight(bonusMetalBaseURL, "/") + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build request: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch %s: %w", path, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %s response: %w", path, err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, nil, errNotFound
	}
	if res.StatusCode >= 400 {
		slog.Warn("Bonus Metal request failed", "path", path, "status", res.StatusCode, "response", string(body))
		return nil, nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}
	return body, res.Header, nil
}
