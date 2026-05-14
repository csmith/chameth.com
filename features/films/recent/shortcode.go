package recent

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func RegisterShortcodes(mgr *shortcodes.Manager) {
	mgr.Register("recentfilms", RenderFromText)
}
