package nextraces

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
	"tailscale.com/tsnet"
)

const (
	shortcodeVersion = 1
	refreshFrequency = 12 * time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"nextraces",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[[]race], error) {
			return retrieve(ctx, ts.HTTPClient(), args)
		},
		render,
	)
}

// A race, trimmed to what the countdown row needs: races are labelled by
// distance alone, so the API's name and type are dropped.
type race struct {
	Date       string  `json:"date"`
	DistanceKm float64 `json:"distance_km"`
}

func retrieve(ctx context.Context, client *http.Client, _ []string) (shortcodes.Result[[]race], error) {
	body, err := centralParkGet(ctx, client, "/api/events", url.Values{"type": {"race"}})
	if err != nil {
		return shortcodes.Result[[]race]{}, fmt.Errorf("failed to fetch races: %w", err)
	}

	var events []struct {
		Date     string  `json:"date"`
		Distance float64 `json:"distance"`
	}
	if err := json.Unmarshal(body, &events); err != nil {
		return shortcodes.Result[[]race]{}, fmt.Errorf("failed to decode events response: %w", err)
	}

	races := make([]race, 0, len(events))
	for _, e := range events {
		// The API omits distance when an event has no category or
		// course distance; an unmeasured race has nothing to be
		// labelled with.
		if e.Distance <= 0 {
			continue
		}
		races = append(races, race{Date: e.Date, DistanceKm: e.Distance})
	}

	return shortcodes.Result[[]race]{
		Data:      races,
		RefreshAt: shortcodes.RefreshIn(refreshFrequency),
	}, nil
}

const centralParkBaseURL = "https://cp.yak-wall.ts.net"

func centralParkGet(ctx context.Context, client *http.Client, path string, query url.Values) ([]byte, error) {
	u := strings.TrimRight(centralParkBaseURL, "/") + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", path, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s response: %w", path, err)
	}

	if res.StatusCode >= 400 {
		slog.Warn("Central Park request failed", "path", path, "status", res.StatusCode, "response", string(body))
		return nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}
	return body, nil
}
