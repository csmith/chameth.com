package common

import (
	"context"

	"chameth.com/chameth.com/features/media"
)

type Context struct {
	context.Context
	Media []media.MediaRelationWithDetails
	URL   string
}

func (c *Context) MediaWithDescription(description string) []media.MediaRelationWithDetails {
	var matching []media.MediaRelationWithDetails
	for i := range c.Media {
		if c.Media[i].Description != nil && *c.Media[i].Description == description {
			matching = append(matching, c.Media[i])
		}
	}
	return matching
}
