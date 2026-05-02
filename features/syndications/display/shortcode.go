package shortcode

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("syndication", RenderFromText)
}
