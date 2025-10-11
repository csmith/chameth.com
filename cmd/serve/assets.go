package main

import (
	"embed"
	"net/http"
)

//go:embed assets/fonts/*
var assets embed.FS

func serveAssets() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, assets, r.URL.Path)
	})
}
