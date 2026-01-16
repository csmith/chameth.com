package shortcodes

import "chameth.com/chameth.com/cmd/serve/db"

type Context struct {
	Media []db.MediaRelationWithDetails
}

func (c *Context) MediaWithDescription(description string) []db.MediaRelationWithDetails {
	var matching []db.MediaRelationWithDetails
	for i := range c.Media {
		if c.Media[i].Description != nil && *c.Media[i].Description == description {
			matching = append(matching, c.Media[i])
		}
	}
	return matching
}

type renderer func(string, *Context) (string, error)

var renderers = []renderer{
	renderSideNote,
	renderUpdate,
	renderWarning,
	renderAudio,
	renderVideo,
	renderFigure,
	renderFilmReview,
	renderFilmReviews,
	renderFilmList,
	renderRecentFilms,
	renderRating,
}

func Render(input string, ctx *Context) (string, error) {
	var res = input
	var err error

	for _, r := range renderers {
		if res, err = r(res, ctx); err != nil {
			return "", err
		}
	}

	return res, nil
}
