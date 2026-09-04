package workouts

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
	"time"
)

const pompeiBandBaseURL = "https://pb.yak-wall.ts.net"

var errNotFound = errors.New("not found")

func windowValues(start, end, group string) url.Values {
	v := url.Values{}
	if start != "" {
		v.Set("start", start)
	}
	if end != "" {
		v.Set("end", end)
	}
	if group != "" {
		v.Set("group", group)
	}
	return v
}

type GroupSummary struct {
	Count         int      `json:"count"`
	DistanceM     float64  `json:"distance_m"`
	DurationS     float64  `json:"duration_s"`
	RunDistanceM  *float64 `json:"run_distance_m"`
	WalkDistanceM *float64 `json:"walk_distance_m"`
}

func ActivitySummary(ctx context.Context, client *http.Client, start, end, group string) (map[string]GroupSummary, error) {
	body, err := pompeiBandGet(ctx, client, "/api/insights/activity-summary", windowValues(start, end, group))
	if err != nil {
		return nil, err
	}

	var res struct {
		Groups map[string]GroupSummary `json:"groups"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode activity-summary response: %w", err)
	}
	return res.Groups, nil
}

type DaySummary struct {
	Date   string                  `json:"date"`
	Groups map[string]GroupSummary `json:"groups"`
}

func ActivityDays(ctx context.Context, client *http.Client, start, end, group string) ([]DaySummary, error) {
	body, err := pompeiBandGet(ctx, client, "/api/insights/activity-days", windowValues(start, end, group))
	if err != nil {
		return nil, err
	}

	var res struct {
		Days []DaySummary `json:"days"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode activity-days response: %w", err)
	}
	return res.Days, nil
}

type DistanceRecord struct {
	Group         string    `json:"group"`
	Name          string    `json:"name"`
	DistanceM     float64   `json:"distance_m"`
	RunDistanceM  *float64  `json:"run_distance_m"`
	WalkDistanceM *float64  `json:"walk_distance_m"`
	DurationS     float64   `json:"duration_s"`
	Date          string    `json:"date"`
	StartTime     time.Time `json:"start_time"`
	ActivityID    string    `json:"activity_id"`
}

func (r *DistanceRecord) RankingDistanceM() float64 {
	switch r.Group {
	case "run":
		if r.RunDistanceM != nil {
			return *r.RunDistanceM
		}
	case "walk":
		if r.WalkDistanceM != nil {
			return *r.WalkDistanceM
		}
	}
	return r.DistanceM
}

func GetDistanceRecord(ctx context.Context, client *http.Client, group, start, end string) (*DistanceRecord, error) {
	path := "/api/insights/distance-records/" + url.PathEscape(group)
	body, err := pompeiBandGet(ctx, client, path, windowValues(start, end, ""))
	if err != nil {
		if errors.Is(err, errNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var record DistanceRecord
	if err := json.Unmarshal(body, &record); err != nil {
		return nil, fmt.Errorf("failed to decode distance-records response: %w", err)
	}
	return &record, nil
}

type PersonalBest struct {
	Group         string    `json:"group"`
	DistanceM     float64   `json:"distance_m"`
	ElapsedS      float64   `json:"elapsed_s"`
	GapElapsedS   float64   `json:"gap_elapsed_s"`
	PaceSPerKm    float64   `json:"pace_s_per_km"`
	GapPaceSPerKm float64   `json:"gap_pace_s_per_km"`
	SpeedKmh      float64   `json:"speed_kmh"`
	Date          string    `json:"date"`
	AchievedAt    time.Time `json:"achieved_at"`
	ActivityID    string    `json:"activity_id"`
}

func PBs(ctx context.Context, client *http.Client, group string) ([]PersonalBest, error) {
	body, err := pompeiBandGet(ctx, client, "/api/insights/pbs", windowValues("", "", group))
	if err != nil {
		return nil, err
	}

	var res struct {
		PBs []PersonalBest `json:"pbs"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode pbs response: %w", err)
	}
	return res.PBs, nil
}

type PBEvent struct {
	Group               string    `json:"group"`
	DistanceM           float64   `json:"distance_m"`
	ElapsedS            float64   `json:"elapsed_s"`
	PreviousElapsedS    *float64  `json:"previous_elapsed_s"`
	GapElapsedS         float64   `json:"gap_elapsed_s"`
	PreviousGapElapsedS *float64  `json:"previous_gap_elapsed_s"`
	Date                string    `json:"date"`
	AchievedAt          time.Time `json:"achieved_at"`
	ActivityID          string    `json:"activity_id"`
}

func PBEvents(ctx context.Context, client *http.Client, start, end string) ([]PBEvent, error) {
	body, err := pompeiBandGet(ctx, client, "/api/insights/pb-events", windowValues(start, end, ""))
	if err != nil {
		return nil, err
	}

	var res struct {
		Events []PBEvent `json:"events"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode pb-events response: %w", err)
	}
	return res.Events, nil
}

type ActivityElevation struct {
	GainM float64 `json:"gain_m"`
	LossM float64 `json:"loss_m"`
}

type Activity struct {
	ID            string             `json:"id"`
	StartTime     time.Time          `json:"start_time"`
	DurationS     float64            `json:"duration_s"`
	DistanceM     float64            `json:"distance_m"`
	WalkDistanceM *float64           `json:"walk_distance_m"`
	Elevation     *ActivityElevation `json:"elevation"`
}

func (a *Activity) WalkingDistanceM() float64 {
	if a.WalkDistanceM != nil {
		return *a.WalkDistanceM
	}
	return a.DistanceM
}

func Activities(ctx context.Context, client *http.Client, group string) ([]Activity, error) {
	body, err := pompeiBandGet(ctx, client, "/api/activities", windowValues("", "", group))
	if err != nil {
		return nil, err
	}

	var activities []Activity
	if err := json.Unmarshal(body, &activities); err != nil {
		return nil, fmt.Errorf("failed to decode activities response: %w", err)
	}
	return activities, nil
}

type MonthSummary struct {
	Month              string   `json:"month"`
	Count              int      `json:"count"`
	DistanceM          float64  `json:"distance_m"`
	DurationS          float64  `json:"duration_s"`
	MaxAverageSpeedMps *float64 `json:"max_average_speed_mps"`
	RunDistanceM       *float64 `json:"run_distance_m"`
	WalkDistanceM      *float64 `json:"walk_distance_m"`
}

func ActivityMonths(ctx context.Context, client *http.Client, group string) ([]MonthSummary, error) {
	body, err := pompeiBandGet(ctx, client, "/api/insights/activity-months", windowValues("", "", group))
	if err != nil {
		return nil, err
	}

	var res struct {
		Months []MonthSummary `json:"months"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode activity-months response: %w", err)
	}
	return res.Months, nil
}

func pompeiBandGet(ctx context.Context, client *http.Client, path string, query url.Values) ([]byte, error) {
	u := strings.TrimRight(pompeiBandBaseURL, "/") + path
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

	if res.StatusCode == http.StatusNotFound {
		return nil, errNotFound
	}
	if res.StatusCode >= 400 {
		slog.Warn("Pompei Band request failed", "path", path, "status", res.StatusCode, "response", string(body))
		return nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}
	return body, nil
}
