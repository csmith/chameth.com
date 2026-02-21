package common

import (
	"context"

	"chameth.com/chameth.com/db"
)

type Context struct {
	context.Context
	Media []db.MediaRelationWithDetails
	URL   string
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
