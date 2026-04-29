package review

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("filmreview", RenderFromText)
}
