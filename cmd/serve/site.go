package main

import (
	"chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/features/shortcodes"
	"net/http"
)

type site struct {
	Assets     *assets.Manager
	Shortcodes *shortcodes.Manager
	Mux        *http.ServeMux
}
