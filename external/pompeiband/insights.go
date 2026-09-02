package pompeiband

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// ErrNotFound is returned when an endpoint reports 404, e.g. a
// distance-records lookup with no matching activities.
var ErrNotFound = errors.New("not found")

// windowValues encodes the shared insights window parameters: start/end
// bound start_time as UTC dates (YYYY-MM-DD) or RFC3339 datetimes, with a
// half-open [start, end) window; group is an exact activity_group match.
// Empty values are omitted.
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

// GroupSummary is one activity group's roll-up over a window. The
// run/walk distance splits are omitted by the API when no activity in the
// group carries them.
type GroupSummary struct {
	Count         int      `json:"count"`
	DistanceM     float64  `json:"distance_m"`
	DurationS     float64  `json:"duration_s"`
	RunDistanceM  *float64 `json:"run_distance_m"`
	WalkDistanceM *float64 `json:"walk_distance_m"`
}

// ActivitySummary returns per-group roll-ups (count, total distance, total
// moving time and the gait-classified distance splits) over the window.
// Groups with no activities are absent from the map.
func (c *Client) ActivitySummary(ctx context.Context, start, end, group string) (map[string]GroupSummary, error) {
	body, err := c.get(ctx, "/api/insights/activity-summary", windowValues(start, end, group))
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

// DaySummary is one UTC day's roll-up within an ActivityDays response.
type DaySummary struct {
	Date   string                  `json:"date"`
	Groups map[string]GroupSummary `json:"groups"`
}

// ActivityDays returns the activity-summary roll-ups bucketed per UTC day
// (the day of start_time), oldest first. Only days with at least one
// matching activity are listed.
func (c *Client) ActivityDays(ctx context.Context, start, end, group string) ([]DaySummary, error) {
	body, err := c.get(ctx, "/api/insights/activity-days", windowValues(start, end, group))
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

// DistanceRecord is the single longest activity of one group in a window:
// the "longest run/ride/walk" record. The ranking distance is the group's
// own: run records rank on RunDistanceM and walk records on WalkDistanceM
// (the gait-classified distance actually run or walked, falling back to
// the recorded total when an activity carries no split), so a run's
// walking breaks don't win the run record; every other group ranks on the
// recorded total.
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

// RankingDistanceM returns the distance the record was ranked on: the
// gait-classified split for foot-based groups (falling back to the
// recorded total when the activity carries no split), otherwise the
// recorded total.
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

// DistanceRecord returns the longest activity within the window for the
// given group, or nil when no activity of the group matches.
func (c *Client) DistanceRecord(ctx context.Context, group, start, end string) (*DistanceRecord, error) {
	path := "/api/insights/distance-records/" + url.PathEscape(group)
	body, err := c.get(ctx, path, windowValues(start, end, ""))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
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

// PersonalBest is the current all-time record for one (group, distance)
// combination. PBs are ranked on grade-adjusted (GAP) time, so GapElapsedS
// is the ranking time and ElapsedS the wall-clock one.
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

// PBs returns the current personal bests for a group, one per distance,
// ordered by distance ascending.
func (c *Client) PBs(ctx context.Context, group string) ([]PersonalBest, error) {
	body, err := c.get(ctx, "/api/insights/pbs", windowValues("", "", group))
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

// PBEvent is a single record-breaking effort: the moment a PB was set, the
// time that did it, and the record it displaced. The previous-* fields are
// nil for an inaugural record; otherwise they carry the superseded
// record's times even when that record was set before the window. The
// window bounds AchievedAt, not the activity's start_time.
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

// PBEvents returns every record-breaking effort achieved within the
// window, chronological.
func (c *Client) PBEvents(ctx context.Context, start, end string) ([]PBEvent, error) {
	body, err := c.get(ctx, "/api/insights/pb-events", windowValues(start, end, ""))
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
