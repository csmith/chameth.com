package contact

import (
	"context"
	"time"

	"chameth.com/chameth.com/db"
)

func messageSeenSince(ctx context.Context, message string, since time.Time) (bool, error) {
	return db.Get[bool](
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM contact_metrics
			WHERE md5(message) = md5($1) AND message = $1 AND created_at > $2
		)
	`,
		message,
		since,
	)
}

func checkPairCounts(ctx context.Context) ([]checkPairCount, error) {
	return db.Select[checkPairCount](
		ctx,
		`
		WITH caught AS (
			SELECT DISTINCT id, jsonb_array_elements_text(checks) AS check_name
			FROM contact_metrics
		)
		SELECT a.check_name AS row_check, b.check_name AS col_check, COUNT(*) AS count
		FROM caught a
		JOIN caught b ON a.id = b.id
		GROUP BY a.check_name, b.check_name
	`,
	)
}
