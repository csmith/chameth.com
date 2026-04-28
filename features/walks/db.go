package walks

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
)

func AllWalks(ctx context.Context) ([]Walk, error) {
	walks, err := db.Select[Walk](ctx, "SELECT id, external_id, start_date, end_date, duration_seconds, distance_km, elevation_gain_meters FROM walks ORDER BY start_date DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to get walks: %w", err)
	}
	return walks, nil
}

func TotalDistance(ctx context.Context) (float64, error) {
	dist, err := db.Get[float64](ctx, "SELECT COALESCE(SUM(distance_km), 0) FROM walks")
	if err != nil {
		return 0, fmt.Errorf("failed to get total walking distance: %w", err)
	}
	return dist, nil
}

func MonthlySpeeds(ctx context.Context) ([]MonthlyWalkingSpeed, error) {
	speeds, err := db.Select[MonthlyWalkingSpeed](ctx, `
		SELECT
			DATE_TRUNC('month', start_date) AS month,
			MAX(distance_km / (duration_seconds / 3600.0)) AS avg_speed_kmh
		FROM walks
		GROUP BY DATE_TRUNC('month', start_date)
		ORDER BY month ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly walking speeds: %w", err)
	}
	return speeds, nil
}

func UpsertWalk(ctx context.Context, walk Walk) error {
	_, err := db.Exec(ctx, `
		INSERT INTO walks (external_id, start_date, end_date, duration_seconds, distance_km, elevation_gain_meters)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (external_id)
		DO UPDATE SET
			start_date = EXCLUDED.start_date,
			end_date = EXCLUDED.end_date,
			duration_seconds = EXCLUDED.duration_seconds,
			distance_km = EXCLUDED.distance_km,
			elevation_gain_meters = EXCLUDED.elevation_gain_meters
	`, walk.ExternalID, walk.StartDate, walk.EndDate, walk.DurationSeconds, walk.DistanceKm, walk.ElevationGainMeters)

	if err != nil {
		return fmt.Errorf("failed to upsert walk: %w", err)
	}

	return nil
}
