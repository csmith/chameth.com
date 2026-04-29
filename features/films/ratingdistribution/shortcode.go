package ratingdistribution

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("filmratingdistribution", RenderFromText)
}
