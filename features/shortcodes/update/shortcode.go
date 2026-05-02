package update

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("update", RenderFromText)
}
