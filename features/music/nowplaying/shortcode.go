package nowplaying

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("nowplaying", RenderFromText)
}
