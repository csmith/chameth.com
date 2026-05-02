package review

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("filmreview", RenderFromText)
}
