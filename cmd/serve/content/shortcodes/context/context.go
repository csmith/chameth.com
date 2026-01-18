package context

import (
	"chameth.com/chameth.com/cmd/serve/db"
)

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
