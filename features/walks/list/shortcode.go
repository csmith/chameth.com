package list

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("walks", RenderFromText)
}
