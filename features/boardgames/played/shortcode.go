package played

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("playedbgs", RenderFromText)
}
