package recent

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("recentfilms", RenderFromText)
}
