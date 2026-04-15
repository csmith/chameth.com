package db

import (
	"context"
	"fmt"
)

func UpsertWalk(ctx context.Context, walk Walk) error {
	_, err := Exec(ctx, `
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
