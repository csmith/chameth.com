package nowplaying

import (
	"chameth.com/chameth.com/features/shortcodes"
)

func RegisterShortcodes(mgr *shortcodes.Manager) {
	mgr.Register("nowplaying", RenderFromText)
}
