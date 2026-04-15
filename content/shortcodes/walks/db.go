package walks

import (
	"context"

	"chameth.com/chameth.com/db"
)

func query(ctx context.Context) ([]db.Walk, error) {
	return db.Select[db.Walk](ctx, "SELECT id, external_id, start_date, end_date, duration_seconds, distance_km, elevation_gain_meters FROM walks ORDER BY start_date DESC")
}
