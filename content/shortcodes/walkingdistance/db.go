package walkingdistance

import (
	"context"

	"chameth.com/chameth.com/db"
)

func query(ctx context.Context) (float64, error) {
	return db.Get[float64](ctx, "SELECT COALESCE(SUM(distance_km), 0) FROM walks")
}
