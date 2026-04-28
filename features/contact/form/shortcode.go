package form

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("contact", RenderFromText)
}
