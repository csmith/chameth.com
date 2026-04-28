package speed

import (
	"chameth.com/chameth.com/content/shortcodes"
)

func init() {
	shortcodes.Register("walkingspeed", RenderFromText)
}
