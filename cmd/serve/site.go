package main

import (
	"chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/features/shortcodes"
	"context"
	"net/http"
	"tailscale.com/tsnet"
)

type site struct {
	Context    context.Context
	Tailscale  *tsnet.Server
	Assets     *assets.Manager
	Shortcodes *shortcodes.Manager
	Mux        *http.ServeMux
}
