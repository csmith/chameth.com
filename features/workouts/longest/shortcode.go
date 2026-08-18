package longest

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func RegisterShortcodes(mgr *shortcodes.Manager) {
	mgr.Register("longestrun", RenderRunFromText)
	mgr.Register("longestcycle", RenderCycleFromText)
}
