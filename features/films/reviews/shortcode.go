package reviews

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("filmreviews", RenderFromText)
}
