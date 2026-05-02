package wowchar

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("wowchar", RenderFromText)
}
