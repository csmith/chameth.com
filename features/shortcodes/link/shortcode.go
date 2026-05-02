package link

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("link", RenderFromText)
}
