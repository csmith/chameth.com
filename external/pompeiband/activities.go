package pompeiband

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
)

type Elevation struct {
	GainM float64 `json:"gain_m"`
	LossM float64 `json:"loss_m"`
}

type Gait struct {
	RunDistanceM      float64 `json:"run_distance_m"`
	WalkDistanceM     float64 `json:"walk_distance_m"`
	OverallPaceSPerKm float64 `json:"overall_pace_s_per_km"`
	OverallCadenceSpm float64 `json:"overall_cadence_spm"`
	RunPaceSPerKm     float64 `json:"run_pace_s_per_km"`
	RunCadenceSpm     float64 `json:"run_cadence_spm"`
	WalkPaceSPerKm    float64 `json:"walk_pace_s_per_km"`
	WalkCadenceSpm    float64 `json:"walk_cadence_spm"`
}

type Weather struct {
	TempC         float64 `json:"temp_c"`
	ApparentTempC float64 `json:"apparent_temp_c"`
	HumidityPct   float64 `json:"humidity_pct"`
	WeatherCode   int     `json:"weather_code"`
	WeatherLabel  string  `json:"weather_label"`
	PrecipMm      float64 `json:"precip_mm"`
	SnowfallCm    float64 `json:"snowfall_cm"`
	CloudCoverPct float64 `json:"cloud_cover_pct"`
	VisibilityM   float64 `json:"visibility_m"`
}

type Segment struct {
	DistanceM     float64   `json:"distance_m"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	ElapsedS      float64   `json:"elapsed_s"`
	PaceSPerKm    float64   `json:"pace_s_per_km"`
	SpeedKmh      float64   `json:"speed_kmh"`
	GapElapsedS   float64   `json:"gap_elapsed_s"`
	GapPaceSPerKm float64   `json:"gap_pace_s_per_km"`
	IsPB          bool      `json:"is_pb"`
	PreviousBestS *float64  `json:"previous_best_s,omitempty"`
	BestS         *float64  `json:"best_s,omitempty"`
	Rank          int       `json:"rank"`
}

// Activity deliberately omits heart_rate_zones and hr_zone_basis, which are
// not being ingested; json.Unmarshal silently ignores unrecognized fields.
type Activity struct {
	ID            string     `json:"id"`
	Activity      string     `json:"activity"`
	ActivityGroup string     `json:"activity_group"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       time.Time  `json:"end_time"`
	DurationS     float64    `json:"duration_s"`
	DistanceM     float64    `json:"distance_m"`
	ActiveS       float64    `json:"active_s"`
	ElapsedS      float64    `json:"elapsed_s"`
	PausedS       float64    `json:"paused_s"`
	Calories      *float64   `json:"calories,omitempty"`
	RunDistanceM  *float64   `json:"run_distance_m,omitempty"`
	WalkDistanceM *float64   `json:"walk_distance_m,omitempty"`
	Elevation     *Elevation `json:"elevation,omitempty"`
	Gait          *Gait      `json:"gait,omitempty"`
	Segments      []Segment  `json:"segments"`
	Weather       *Weather   `json:"weather,omitempty"`
}

// GetActivities fetches activities with start_time strictly after since.
// Pass since == "" to fetch full history.
func (c *Client) GetActivities(ctx context.Context, since string) ([]Activity, error) {
	u := strings.TrimRight(c.baseURL, "/") + "/api/activities"
	if since != "" {
		q := url.Values{}
		q.Set("since", since)
		u += "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	res, err := c.h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch activities: %w", err)
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read activities response: %w", err)
	}

	if res.StatusCode >= 400 {
		slog.Warn("Pompei Band request failed", "status", res.StatusCode, "response", string(b))
		return nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}

	var activities []Activity
	if err := json.Unmarshal(b, &activities); err != nil {
		return nil, fmt.Errorf("failed to decode activities response: %w", err)
	}
	return activities, nil
}
