package figure

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("figure", RenderFromText)
}
