package boardgames

import (
	"context"
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
	res, err := twowords.GetJSON[struct {
		Games []Game `json:"games"`
	}](ctx, magicMeters, client, "/api/playcounts", twowords.Query{"start": start, "end": end})
	if err != nil {
		return nil, err
	}
	return res.Games, nil
}
