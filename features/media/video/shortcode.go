package video

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("video", RenderFromText)
}
