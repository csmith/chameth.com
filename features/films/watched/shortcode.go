package watched

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("watchedfilms", RenderFromText)
}
