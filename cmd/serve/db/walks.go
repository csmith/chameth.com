package db

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/cmd/serve/metrics"
)

// UpsertWalk inserts or updates a walk based on external_id
func UpsertWalk(ctx context.Context, walk Walk) error {
	metrics.LogQuery(ctx)

	_, err := db.ExecContext(ctx, `
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

// GetAllWalks retrieves all walks ordered by start_date descending
func GetAllWalks(ctx context.Context) ([]Walk, error) {
	metrics.LogQuery(ctx)
	var walks []Walk
	err := db.SelectContext(ctx, &walks, "SELECT id, external_id, start_date, end_date, duration_seconds, distance_km, elevation_gain_meters FROM walks ORDER BY start_date DESC")
	if err != nil {
		return nil, err
	}
	return walks, nil
}

// GetTotalDistanceKm returns the total distance walked in kilometers
func GetTotalDistanceKm(ctx context.Context) (float64, error) {
	metrics.LogQuery(ctx)
	var totalDistance float64
	err := db.GetContext(ctx, &totalDistance, "SELECT COALESCE(SUM(distance_km), 0) FROM walks")
	if err != nil {
		return 0, err
	}
	return totalDistance, nil
}
