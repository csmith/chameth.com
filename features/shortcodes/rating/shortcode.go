package rating

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("rating", RenderFromText)
}
