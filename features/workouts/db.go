package workouts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"chameth.com/chameth.com/db"
)

// cursorSince returns the most recent workout start_time, formatted for use
// as the activity tracker API's `since` query param, or "" if no workouts
// have been synced yet (triggering a full-history fetch).
func cursorSince(ctx context.Context) (string, error) {
	t, err := db.Get[*time.Time](ctx, `SELECT MAX(start_time) FROM workouts`)
	if err != nil {
		return "", fmt.Errorf("failed to get workout sync cursor: %w", err)
	}
	if t == nil {
		return "", nil
	}
	return t.Format(time.RFC3339), nil
}

// TotalsInRange returns per-activity workout counts, durations and
// distances for workouts starting within [start, end].
func TotalsInRange(ctx context.Context, start, end time.Time) (PeriodTotals, error) {
	return db.Get[PeriodTotals](ctx, `
		SELECT
			COUNT(*) FILTER (WHERE activity_group = 'cycle') AS cycle_count,
			COALESCE(SUM(distance_m) FILTER (WHERE activity_group = 'cycle'), 0) AS cycle_distance_m,
			COALESCE(SUM(duration_s) FILTER (WHERE activity_group = 'cycle'), 0) AS cycle_duration_s,
			COUNT(*) FILTER (WHERE activity_group = 'run') AS run_count,
			COALESCE(SUM(run_distance_m), 0) AS run_distance_m,
			COALESCE(SUM(duration_s) FILTER (WHERE activity_group = 'run'), 0) AS run_duration_s
		FROM workouts
		WHERE start_time >= $1 AND start_time <= $2
	`, start, end)
}

// FurthestCycleInRange returns the longest cycle workout starting within
// [start, end], or nil if there were none.
func FurthestCycleInRange(ctx context.Context, start, end time.Time) (*FurthestWorkout, error) {
	w, err := db.Get[FurthestWorkout](ctx, `
		SELECT activity, distance_m, start_time
		FROM workouts
		WHERE activity_group = 'cycle' AND start_time >= $1 AND start_time <= $2
		ORDER BY distance_m DESC
		LIMIT 1
	`, start, end)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get furthest cycle: %w", err)
	}
	return &w, nil
}

// FurthestRunInRange returns the longest running interval starting within
// [start, end], or nil if there were none. This is deliberately not
// filtered to activity_group = 'run', since running intervals can occur
// inside 'walk'-grouped workouts too (and vice versa).
func FurthestRunInRange(ctx context.Context, start, end time.Time) (*FurthestWorkout, error) {
	w, err := db.Get[FurthestWorkout](ctx, `
		SELECT activity, run_distance_m AS distance_m, start_time
		FROM workouts
		WHERE run_distance_m IS NOT NULL AND start_time >= $1 AND start_time <= $2
		ORDER BY run_distance_m DESC
		LIMIT 1
	`, start, end)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get furthest run: %w", err)
	}
	return &w, nil
}

// FurthestCycle returns the longest cycle workout ever recorded, or nil if
// there were none.
func FurthestCycle(ctx context.Context) (*FurthestWorkout, error) {
	w, err := db.Get[FurthestWorkout](ctx, `
		SELECT activity, distance_m, start_time
		FROM workouts
		WHERE activity_group = 'cycle'
		ORDER BY distance_m DESC
		LIMIT 1
	`)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get furthest cycle: %w", err)
	}
	return &w, nil
}

// FurthestRun returns the longest running interval ever recorded, or nil if
// there were none. Like FurthestRunInRange this is deliberately not filtered
// to activity_group = 'run', since running intervals can occur inside
// 'walk'-grouped workouts too.
func FurthestRun(ctx context.Context) (*FurthestWorkout, error) {
	w, err := db.Get[FurthestWorkout](ctx, `
		SELECT activity, run_distance_m AS distance_m, start_time
		FROM workouts
		WHERE run_distance_m IS NOT NULL
		ORDER BY run_distance_m DESC
		LIMIT 1
	`)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get furthest run: %w", err)
	}
	return &w, nil
}

// RecentPBsInRange returns, for each cycle/run segment-distance
// combination, the fastest (and therefore most recent) PB set within
// [start, end]. Other activity groups (e.g. walking) are excluded.
// PreviousBestS is taken from the earliest PB set within the range, so it
// reflects the record that stood before the period began, rather than the
// time it most recently improved on (if it was beaten more than once
// within the range).
func RecentPBsInRange(ctx context.Context, start, end time.Time) ([]PersonalBest, error) {
	return db.Select[PersonalBest](ctx, `
		SELECT DISTINCT ON (w.activity_group, ws.distance_m)
			w.activity_group, w.activity, ws.distance_m, ws.elapsed_s, ws.start_time,
			FIRST_VALUE(ws.previous_best_s) OVER (
				PARTITION BY w.activity_group, ws.distance_m ORDER BY ws.start_time ASC
			) AS previous_best_s
		FROM workout_segments ws
		JOIN workouts w ON w.id = ws.workout_id
		WHERE ws.is_pb = true AND w.activity_group IN ('cycle', 'run')
			AND ws.start_time >= $1 AND ws.start_time <= $2
		ORDER BY w.activity_group, ws.distance_m, ws.elapsed_s ASC
	`, start, end)
}

// insertWorkoutWithSegments inserts a workout and its segments in a single
// transaction. If a workout with the same external_id already exists, this
// is a no-op rather than an error.
func insertWorkoutWithSegments(ctx context.Context, w workout, segments []workoutSegment) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var workoutID int
	err = tx.QueryRow(`
		INSERT INTO workouts (
			external_id, activity, activity_group, start_time, end_time,
			duration_s, distance_m, active_s, elapsed_s, paused_s,
			calories, run_distance_m, walk_distance_m,
			elevation_gain_m, elevation_loss_m,
			gait_run_distance_m, gait_walk_distance_m, gait_overall_pace_s_per_km, gait_overall_cadence_spm,
			gait_run_pace_s_per_km, gait_run_cadence_spm, gait_walk_pace_s_per_km, gait_walk_cadence_spm,
			weather_temp_c, weather_apparent_temp_c, weather_humidity_pct, weather_code, weather_label,
			weather_precip_mm, weather_snowfall_cm, weather_cloud_cover_pct, weather_visibility_m
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13,
			$14, $15,
			$16, $17, $18, $19,
			$20, $21, $22, $23,
			$24, $25, $26, $27, $28,
			$29, $30, $31, $32
		)
		ON CONFLICT (external_id) DO NOTHING
		RETURNING id
	`,
		w.ExternalID, w.Activity, w.ActivityGroup, w.StartTime, w.EndTime,
		w.DurationS, w.DistanceM, w.ActiveS, w.ElapsedS, w.PausedS,
		w.Calories, w.RunDistanceM, w.WalkDistanceM,
		w.ElevationGainM, w.ElevationLossM,
		w.GaitRunDistanceM, w.GaitWalkDistanceM, w.GaitOverallPaceSPerKm, w.GaitOverallCadenceSpm,
		w.GaitRunPaceSPerKm, w.GaitRunCadenceSpm, w.GaitWalkPaceSPerKm, w.GaitWalkCadenceSpm,
		w.WeatherTempC, w.WeatherApparentTempC, w.WeatherHumidityPct, w.WeatherCode, w.WeatherLabel,
		w.WeatherPrecipMm, w.WeatherSnowfallCm, w.WeatherCloudCoverPct, w.WeatherVisibilityM,
	).Scan(&workoutID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to insert workout: %w", err)
	}

	if err == nil {
		for _, seg := range segments {
			_, err = tx.Exec(`
				INSERT INTO workout_segments (
					workout_id, segment_index, distance_m, start_time, end_time,
					elapsed_s, pace_s_per_km, speed_kmh, gap_elapsed_s, gap_pace_s_per_km,
					is_pb, previous_best_s, best_s, rank
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			`,
				workoutID, seg.SegmentIndex, seg.DistanceM, seg.StartTime, seg.EndTime,
				seg.ElapsedS, seg.PaceSPerKm, seg.SpeedKmh, seg.GapElapsedS, seg.GapPaceSPerKm,
				seg.IsPb, seg.PreviousBestS, seg.BestS, seg.Rank,
			)
			if err != nil {
				return fmt.Errorf("failed to insert workout segment %d: %w", seg.SegmentIndex, err)
			}
		}
	} else {
		// Workout already existed (ON CONFLICT DO NOTHING); nothing to insert.
		err = nil
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
