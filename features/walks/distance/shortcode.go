package distance

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func init() {
	shortcodes.Register("walkingdistance", RenderFromText)
}
