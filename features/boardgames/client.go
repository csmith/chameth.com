package boardgames

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"chameth.com/chameth.com/external/twowords"
)

var magicMeters = twowords.Service{Name: "Magic Meters", BaseURL: "https://mm.yak-wall.ts.net"}

type Game struct {
	ID         string `json:"id"`
	BggID      *int   `json:"bgg_id"`
	Name       string `json:"name"`
	Year       *int   `json:"year"`
	ImageURL   string `json:"image_url"`
	PlayCount  int    `json:"play_count"`
	LastPlayed string `json:"last_played"`
}

// start and end define a half-open range; empty values mean all time.
func PlayCounts(ctx context.Context, client *http.Client, start, end string) ([]Game, error) {
	body, _, err := magicMeters.Get(ctx, client, "/api/playcounts", twowords.Query{"start": start, "end": end})
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
	body, header, err := magicMeters.Get(ctx, client, twowords.Path("/games/%s/image", gameID), nil)
	if err != nil {
		return nil, "", err
	}
	return body, header.Get("Content-Type"), nil
}
