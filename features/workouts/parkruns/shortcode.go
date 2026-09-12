// Package parkruns renders shortcodes backed by the Central Park
// parkruns insight API.
package parkruns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
	"tailscale.com/tsnet"
)

const (
	centralParkBaseURL = "https://cp.yak-wall.ts.net"
	refreshFrequency   = 12 * time.Hour

	parkrunstatsVersion = 2
	parkrunsVersion     = 2
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"parkrunstats",
		parkrunstatsVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[statsData], error) {
			return retrieveStats(ctx, ts.HTTPClient(), args)
		},
		renderStats,
	)
	mgr.RegisterData(
		"parkruns",
		parkrunsVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[[]runRecord], error) {
			return retrieveTable(ctx, ts.HTTPClient(), args)
		},
		renderTable,
	)
}

// bestRun is the fastest recorded gun time. Date and Location are
// formatted for display; Time is the gun time as recorded.
type bestRun struct {
	Time     string `json:"time"`
	Date     string `json:"date"`
	Location string `json:"location"`
}

type statsData struct {
	Total       int      `json:"total"`
	Venues      int      `json:"venues"`
	Best        *bestRun `json:"best"`
	LatestDate  string   `json:"latest_date"`
	LatestVenue string   `json:"latest_venue"`
}

func retrieveStats(ctx context.Context, client *http.Client, _ []string) (shortcodes.Result[statsData], error) {
	res, err := fetchParkruns(ctx, client)
	if err != nil {
		return shortcodes.Result[statsData]{}, fmt.Errorf("failed to fetch parkruns: %w", err)
	}

	d := statsData{Total: res.Total, Venues: res.Venues}
	for _, run := range res.Runs {
		if run.Date >= d.LatestDate {
			d.LatestDate = run.Date
			d.LatestVenue = run.Location
		}
	}
	if res.Best != nil {
		d.Best = &bestRun{Time: res.Best.Time, Date: res.Best.Date, Location: res.Best.Location}
	}

	return shortcodes.Result[statsData]{
		Data:      d,
		RefreshAt: shortcodes.RefreshIn(refreshFrequency),
	}, nil
}

// runRecord is a single completed parkrun. It is the cached data shape,
// so it must round-trip through JSON.
type runRecord struct {
	Date       string `json:"date"`
	Name       string `json:"name"`
	Location   string `json:"location"`
	GunTime    string `json:"gun_time"`
	ChipTime   string `json:"chip_time"`
	PosOverall *int   `json:"rank_overall"`
	PosAge     *int   `json:"rank_age_group"`
}

func parseRange(args []string) (start, end time.Time, err error) {
	if len(args) == 0 {
		return time.Time{}, time.Time{}, nil
	}
	if len(args) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("parkruns requires 0 or 2 arguments (start_date, end_date) in YYYY-MM-DD format")
	}

	start, err = time.Parse("2006-01-02", args[0])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %s (expected YYYY-MM-DD)", args[0])
	}

	end, err = time.Parse("2006-01-02", args[1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %s (expected YYYY-MM-DD)", args[1])
	}

	if start.After(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("start date must not be after end date")
	}

	return start, end, nil
}

func retrieveTable(ctx context.Context, client *http.Client, args []string) (shortcodes.Result[[]runRecord], error) {
	start, end, err := parseRange(args)
	if err != nil {
		return shortcodes.Result[[]runRecord]{}, err
	}

	res, err := fetchParkruns(ctx, client)
	if err != nil {
		return shortcodes.Result[[]runRecord]{}, fmt.Errorf("failed to fetch parkruns: %w", err)
	}

	runs := make([]runRecord, 0, len(res.Runs))
	for _, r := range res.Runs {
		date, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			continue
		}
		if !start.IsZero() && (date.Before(start) || date.After(end)) {
			continue
		}
		runs = append(runs, r)
	}

	return shortcodes.Result[[]runRecord]{
		Data:      runs,
		RefreshAt: shortcodes.RefreshIn(refreshFrequency),
	}, nil
}

type parkrunsResponse struct {
	Total  int         `json:"total"`
	Venues int         `json:"venues"`
	Best   *bestRun    `json:"best"`
	Runs   []runRecord `json:"runs"`
}

func fetchParkruns(ctx context.Context, client *http.Client) (*parkrunsResponse, error) {
	body, err := centralParkGet(ctx, client, "/api/insights/parkruns")
	if err != nil {
		return nil, err
	}

	var res parkrunsResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode parkruns response: %w", err)
	}
	return &res, nil
}

func centralParkGet(ctx context.Context, client *http.Client, path string) ([]byte, error) {
	u := strings.TrimRight(centralParkBaseURL, "/") + path

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
