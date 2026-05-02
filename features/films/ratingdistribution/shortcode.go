package ratingdistribution

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("filmratingdistribution", RenderFromText)
}
