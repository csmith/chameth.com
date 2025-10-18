package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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

	if err := initDatabase(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	if err := updateStylesheet(); err != nil {
		slog.Error("Failed to update stylesheet", "error", err)
		os.Exit(1)
	}

	go func() {
		UpdateAllPostEmbeddings(context.Background())
	}()

	mux := http.NewServeMux()
	mux.Handle("POST /api/contact", http.HandlerFunc(handleContactForm))
	mux.Handle("GET /assets/stylesheets/", serveStylesheet())
	mux.Handle("GET /index.xml", http.HandlerFunc(handleFullFeed))
	mux.Handle("GET /short.xml", http.HandlerFunc(handleShortPostsFeed))
	mux.Handle("GET /long.xml", http.HandlerFunc(handleLongPostsFeed))
	mux.Handle("GET /posts/{$}", http.HandlerFunc(handlePostsList))
	mux.Handle("GET /prints/{$}", http.HandlerFunc(handlePrintsList))
	mux.Handle("GET /projects/{$}", http.HandlerFunc(handleProjectsList))
	mux.Handle("GET /sitemap/{$}", http.HandlerFunc(handleSiteMap))
	mux.Handle("GET /snippets/{$}", http.HandlerFunc(handleSnippetsList))
	mux.Handle("GET /{$}", http.HandlerFunc(handleAbout))
	mux.Handle("/", http.HandlerFunc(handleContent))

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
				// middleware.Recover(),
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
