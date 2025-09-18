package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/csmith/envflag/v2"
	"github.com/csmith/middleware"
	"github.com/csmith/slogflags"
)

var (
	port  = flag.Int("port", 8080, "Port to listen on")
	files = flag.String("files", "_site", "Directory to serve files from")
)

func main() {
	envflag.Parse()
	_ = slogflags.Logger(slogflags.WithSetDefault(true))

	mux := http.NewServeMux()
	mux.Handle("POST /api/contact", http.HandlerFunc(handleContactForm))
	mux.Handle("/", http.FileServer(http.Dir(*files)))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler: middleware.Chain(
			middleware.WithMiddleware(
				middleware.RealAddress(),
				middleware.CrossOriginProtection(),
				middleware.ErrorHandler(
					middleware.WithErrorHandler(http.StatusNotFound, http.HandlerFunc(handleNotFound)),
				),
				middleware.CacheControl(
					middleware.WithCacheTimes(map[string]time.Duration{
						"application/*":    time.Hour * 24 * 365,
						"application/xml":  time.Hour,
						"application/json": time.Duration(0),
						"audio/*":          time.Hour * 24 * 365,
						"font/*":           time.Hour * 24 * 365,
						"image/*":          time.Hour * 24 * 365,
						"text/*":           time.Hour,
						"text/css":         time.Hour * 24 * 365,
						"video/*":          time.Hour * 24 * 365,
					}),
				),
				middleware.Compress(),
				middleware.Headers(
					middleware.WithHeader("X-Content-Type-Options", "nosniff"),
					middleware.WithHeader("Content-Security-Policy", "default-src 'self' https://chameth.com/ https://contact.chameth.com https://u.c5h.io/ 'nonce-littlefoot-ae805b14';"),
					middleware.WithHeader("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload"),
					middleware.WithHeader("Referrer-Policy", "no-referrer-when-downgrade"),
				),
				middleware.Recover(),
				applyRedirects(),
			),
		)(mux),
	}

	go func() {
		slog.Info(fmt.Sprintf("Listening on http://0.0.0.0:%d/", *port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to listen", "error", err)
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown HTTP server", "error", err)
		panic(err)
	}
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	f, err := os.Open(filepath.Join(*files, "404.html"))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	io.Copy(w, f)
}
