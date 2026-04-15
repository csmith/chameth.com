package filmratingdistribution

import (
	"context"

	"chameth.com/chameth.com/db"
)

func query(ctx context.Context) ([]db.FilmRatingDistribution, error) {
	return db.Select[db.FilmRatingDistribution](ctx, `
		SELECT rating, COUNT(*) as count
		FROM film_reviews
		WHERE published = true
		GROUP BY rating
		ORDER BY rating ASC
	`)
}
