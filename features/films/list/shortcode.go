package list

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("filmlist", RenderFromText)
}
