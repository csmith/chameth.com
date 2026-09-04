package boardgames

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

const magicMetersBaseURL = "https://mm.yak-wall.ts.net"

var errNotFound = errors.New("not found")

// Game is one game's rolled-up play data. Year and ImageURL are omitted by
// the API when the year is unknown or the box art has not been fetched yet.
type Game struct {
	ID         string `json:"id"`
	BggID      *int   `json:"bgg_id"`
	Name       string `json:"name"`
	Year       *int   `json:"year"`
	ImageURL   string `json:"image_url"`
	PlayCount  int    `json:"play_count"`
	LastPlayed string `json:"last_played"`
}

// PlayCounts counts plays per game over a window. The window is half-open:
// start is inclusive, end exclusive, both YYYY-MM-DD dates; both empty
// means all time.
func PlayCounts(ctx context.Context, client *http.Client, start, end string) ([]Game, error) {
	var query url.Values
	if start != "" || end != "" {
		query = url.Values{}
		query.Set("start", start)
		query.Set("end", end)
	}

	body, _, err := magicMetersGet(ctx, client, "/api/playcounts", query)
	if err != nil {
		return nil, err
	}

	var res struct {
		Games []Game `json:"games"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode playcounts response: %w", err)
	}
	return res.Games, nil
}

func fetchImage(ctx context.Context, client *http.Client, gameID string) ([]byte, string, error) {
	path := "/games/" + url.PathEscape(gameID) + "/image"
	body, header, err := magicMetersGet(ctx, client, path, nil)
	if err != nil {
		return nil, "", err
	}
	return body, header.Get("Content-Type"), nil
}

func magicMetersGet(ctx context.Context, client *http.Client, path string, query url.Values) ([]byte, http.Header, error) {
	u := strings.TrimRight(magicMetersBaseURL, "/") + path
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
		slog.Warn("Magic Meters request failed", "path", path, "status", res.StatusCode, "response", string(body))
		return nil, nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}
	return body, res.Header, nil
}
