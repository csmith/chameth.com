package distance

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("walkingdistance", RenderFromText)
}
