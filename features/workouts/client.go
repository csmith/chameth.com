package workouts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"chameth.com/chameth.com/external/twowords"
)

var pompeiBand = twowords.Service{Name: "Pompei Band", BaseURL: "https://pb.yak-wall.ts.net"}

// CentralPark serves race and parkrun data.
var CentralPark = twowords.Service{Name: "Central Park", BaseURL: "https://cp.yak-wall.ts.net"}

type GroupSummary struct {
	Count         int      `json:"count"`
	DistanceM     float64  `json:"distance_m"`
	DurationS     float64  `json:"duration_s"`
	RunDistanceM  *float64 `json:"run_distance_m"`
	WalkDistanceM *float64 `json:"walk_distance_m"`
}

func ActivitySummary(
	ctx context.Context, client *http.Client, start, end, group string,
) (map[string]GroupSummary, error) {
	body, _, err := pompeiBand.Get(
		ctx,
		client,
		"/api/insights/activity-summary",
		twowords.Query{"start": start, "end": end, "group": group},
	)
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
	body, _, err := pompeiBand.Get(
		ctx,
		client,
		"/api/insights/activity-days",
		twowords.Query{"start": start, "end": end, "group": group},
	)
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
	body, _, err := pompeiBand.Get(
		ctx,
		client,
		twowords.Path("/api/insights/distance-records/%s", group),
		twowords.Query{"start": start, "end": end},
	)
	if err != nil {
		if errors.Is(err, twowords.ErrNotFound) {
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
	body, _, err := pompeiBand.Get(ctx, client, "/api/insights/pbs", twowords.Query{"group": group})
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
	body, _, err := pompeiBand.Get(ctx, client, "/api/insights/pb-events", twowords.Query{"start": start, "end": end})
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
	body, _, err := pompeiBand.Get(ctx, client, "/api/activities", twowords.Query{"group": group})
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
	body, _, err := pompeiBand.Get(ctx, client, "/api/insights/activity-months", twowords.Query{"group": group})
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
