package watched

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("watchedfilms", RenderFromText)
}
