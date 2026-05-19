package main

import (
	"chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/features/routing"
	"chameth.com/chameth.com/features/shortcodes"
	"context"
	"tailscale.com/tsnet"
)

type site struct {
	Context    context.Context
	Tailscale  *tsnet.Server
	Assets     *assets.Manager
	Shortcodes *shortcodes.Manager
	Routes     *routing.Manager
}
