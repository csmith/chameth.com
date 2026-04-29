package link

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("postlink", RenderFromText)
}
