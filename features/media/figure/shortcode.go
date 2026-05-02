package figure

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("figure", RenderFromText)
}
