package wowchar

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("wowchar", RenderFromText)
}
