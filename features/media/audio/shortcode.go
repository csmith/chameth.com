package audio

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("audio", RenderFromText)
}
