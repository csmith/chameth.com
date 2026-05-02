package topalbums

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("topalbums", RenderFromText)
}
