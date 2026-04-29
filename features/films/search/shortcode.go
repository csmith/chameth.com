package search

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("filmsearch", RenderFromText)
}
