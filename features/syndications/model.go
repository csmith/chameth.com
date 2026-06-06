package syndications

import "chameth.com/chameth.com/features/posts"

type Syndication struct {
	ID          int     `db:"id"`
	Path        string  `db:"path"`
	ExternalURL string  `db:"external_url"`
	Name        string  `db:"name"`
	Published   bool    `db:"published"`
	Disposition string  `db:"disposition"`
	Rel         *string `db:"rel"`
}

type blueskySyndicationWithPost struct {
	Syndication
	posts.PostMetadata
}
