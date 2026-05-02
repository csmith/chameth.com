package topartists

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("topartists", RenderFromText)
}
