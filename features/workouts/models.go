package workouts

import "time"

type workout struct {
	ID                    int       `db:"id"`
	ExternalID            string    `db:"external_id"`
	Activity              string    `db:"activity"`
	ActivityGroup         string    `db:"activity_group"`
	StartTime             time.Time `db:"start_time"`
	EndTime               time.Time `db:"end_time"`
	DurationS             float64   `db:"duration_s"`
	DistanceM             float64   `db:"distance_m"`
	ActiveS               float64   `db:"active_s"`
	ElapsedS              float64   `db:"elapsed_s"`
	PausedS               float64   `db:"paused_s"`
	Calories              *float64  `db:"calories"`
	RunDistanceM          *float64  `db:"run_distance_m"`
	WalkDistanceM         *float64  `db:"walk_distance_m"`
	ElevationGainM        *float64  `db:"elevation_gain_m"`
	ElevationLossM        *float64  `db:"elevation_loss_m"`
	GaitRunDistanceM      *float64  `db:"gait_run_distance_m"`
	GaitWalkDistanceM     *float64  `db:"gait_walk_distance_m"`
	GaitOverallPaceSPerKm *float64  `db:"gait_overall_pace_s_per_km"`
	GaitOverallCadenceSpm *float64  `db:"gait_overall_cadence_spm"`
	GaitRunPaceSPerKm     *float64  `db:"gait_run_pace_s_per_km"`
	GaitRunCadenceSpm     *float64  `db:"gait_run_cadence_spm"`
	GaitWalkPaceSPerKm    *float64  `db:"gait_walk_pace_s_per_km"`
	GaitWalkCadenceSpm    *float64  `db:"gait_walk_cadence_spm"`
	WeatherTempC          *float64  `db:"weather_temp_c"`
	WeatherApparentTempC  *float64  `db:"weather_apparent_temp_c"`
	WeatherHumidityPct    *float64  `db:"weather_humidity_pct"`
	WeatherCode           *int      `db:"weather_code"`
	WeatherLabel          *string   `db:"weather_label"`
	WeatherPrecipMm       *float64  `db:"weather_precip_mm"`
	WeatherSnowfallCm     *float64  `db:"weather_snowfall_cm"`
	WeatherCloudCoverPct  *float64  `db:"weather_cloud_cover_pct"`
	WeatherVisibilityM    *float64  `db:"weather_visibility_m"`
}

type workoutSegment struct {
	ID            int       `db:"id"`
	WorkoutID     int       `db:"workout_id"`
	SegmentIndex  int       `db:"segment_index"`
	DistanceM     float64   `db:"distance_m"`
	StartTime     time.Time `db:"start_time"`
	EndTime       time.Time `db:"end_time"`
	ElapsedS      float64   `db:"elapsed_s"`
	PaceSPerKm    float64   `db:"pace_s_per_km"`
	SpeedKmh      float64   `db:"speed_kmh"`
	GapElapsedS   float64   `db:"gap_elapsed_s"`
	GapPaceSPerKm float64   `db:"gap_pace_s_per_km"`
	IsPb          bool      `db:"is_pb"`
	PreviousBestS *float64  `db:"previous_best_s"`
	BestS         *float64  `db:"best_s"`
	Rank          int       `db:"rank"`
}

// PeriodTotals holds per-activity workout counts and distances for a date
// range. Running distance comes from the dedicated run interval column
// rather than a workout's total distance, since a single workout (e.g. a
// couch-to-5k session) can mix running and walking intervals.
type PeriodTotals struct {
	CycleCount     int     `db:"cycle_count"`
	CycleDistanceM float64 `db:"cycle_distance_m"`
	CycleDurationS float64 `db:"cycle_duration_s"`
	RunCount       int     `db:"run_count"`
	RunDistanceM   float64 `db:"run_distance_m"`
	RunDurationS   float64 `db:"run_duration_s"`
}

// WorkoutDayEntry is one workout's contribution to a calendar day: when
// it started, its activity group, the distance covered, and the running
// portion of that distance (nil for workouts with no running intervals).
type WorkoutDayEntry struct {
	StartTime     time.Time `db:"start_time"`
	ActivityGroup string    `db:"activity_group"`
	DistanceM     float64   `db:"distance_m"`
	RunDistanceM  *float64  `db:"run_distance_m"`
}

// FurthestWorkout describes the single longest workout (or run/walk
// interval within a workout) for a date range.
type FurthestWorkout struct {
	Activity  string    `db:"activity"`
	DistanceM float64   `db:"distance_m"`
	StartTime time.Time `db:"start_time"`
}

// PersonalBest describes a segment-distance PB set within a date range.
type PersonalBest struct {
	ActivityGroup string    `db:"activity_group"`
	Activity      string    `db:"activity"`
	DistanceM     float64   `db:"distance_m"`
	ElapsedS      float64   `db:"elapsed_s"`
	StartTime     time.Time `db:"start_time"`
	PreviousBestS *float64  `db:"previous_best_s"`
}
