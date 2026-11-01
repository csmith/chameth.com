package routing

import (
	"context"

	"chameth.com/chameth.com/db"
)

func findExactContentType(ctx context.Context, path string) (string, error) {
	return db.Get[string](ctx, `SELECT content_type FROM paths WHERE path = $1 AND NOT prefix_match`, path)
}
