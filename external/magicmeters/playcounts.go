package magicmeters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

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
// means all time. Games with no play in the window are absent; expansions
// are excluded, matching every statistic the service aggregates. Games are
// ordered by play count descending, then name — an order a client can rely
// on.
func (c *Client) PlayCounts(ctx context.Context, start, end string) ([]Game, error) {
	var query url.Values
	if start != "" || end != "" {
		query = url.Values{}
		query.Set("start", start)
		query.Set("end", end)
	}

	body, _, err := c.get(ctx, "/api/playcounts", query)
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

// Image fetches a game's box art, returning the raw JPEG bytes alongside
// the response's Content-Type. The art is served resized for use as-is
// (short side at most 500px, JPEG quality 85) and is write-once, so the
// bytes behind a game's art never change.
func (c *Client) Image(ctx context.Context, gameID string) ([]byte, string, error) {
	path := "/games/" + url.PathEscape(gameID) + "/image"
	body, header, err := c.get(ctx, path, nil)
	if err != nil {
		return nil, "", err
	}
	return body, header.Get("Content-Type"), nil
}
