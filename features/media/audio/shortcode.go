package audio

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("audio", RenderFromText)
}
