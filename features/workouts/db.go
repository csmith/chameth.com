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
