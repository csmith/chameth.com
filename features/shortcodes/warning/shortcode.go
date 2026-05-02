package warning

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("warning", RenderFromText)
}
