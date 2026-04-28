package played

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("playedbgs", RenderFromText)
}
