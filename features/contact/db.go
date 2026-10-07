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
