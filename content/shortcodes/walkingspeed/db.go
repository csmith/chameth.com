package walkingspeed

import (
	"context"

	"chameth.com/chameth.com/db"
)

func query(ctx context.Context) ([]db.MonthlyWalkingSpeed, error) {
	return db.Select[db.MonthlyWalkingSpeed](ctx, `
		SELECT
			DATE_TRUNC('month', start_date) AS month,
			MAX(distance_km / (duration_seconds / 3600.0)) AS avg_speed_kmh
		FROM walks
		GROUP BY DATE_TRUNC('month', start_date)
		ORDER BY month ASC
	`)
}
