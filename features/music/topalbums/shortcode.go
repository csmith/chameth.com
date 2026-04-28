package topalbums

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("topalbums", RenderFromText)
}
