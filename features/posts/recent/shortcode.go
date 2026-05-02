package recent

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("recentposts", RenderFromText)
}
