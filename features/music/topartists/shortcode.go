package topartists

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("topartists", RenderFromText)
}
