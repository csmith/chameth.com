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

	"chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/features/metrics"
	"chameth.com/chameth.com/features/posts"
	"chameth.com/chameth.com/features/routing"
	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/sudo"
	"chameth.com/chameth.com/handlers"
	"github.com/csmith/envflag/v2"
	"github.com/csmith/middleware"
	"github.com/csmith/slogflags"
	"tailscale.com/tsnet"
)

var (
	port          = flag.Int("port", 8080, "Port to listen on")
	tailscaleHost = flag.String("tailscale-host", "website-admin", "Tailscale host")
	tailscaleDir  = flag.String("tailscale-dir", "tsdata", "Tailscale directory")
)

func main() {
	envflag.Parse()
	_ = slogflags.Logger(slogflags.WithSetDefault(true))

	if err := db.Init(metrics.LogQuery); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := &site{
		Context: ctx,
		Tailscale: &tsnet.Server{
			Hostname: *tailscaleHost,
			Dir:      *tailscaleDir,
			UserLogf: func(s string, v ...any) {
				slog.Info(fmt.Sprintf(s, v...), "source", "tailscale")
			},
			Logf: func(s string, v ...any) {
				slog.Debug(fmt.Sprintf(s, v...), "source", "tailscale")
			},
		},
		Assets:     assets.NewManager(),
		Shortcodes: shortcodes.NewManager(),
		Routes:     routing.NewManager(),
	}

	content.AssetsManager = s.Assets
	content.RecentPostsProvider = posts.Recent
	content.ShortcodesManager = s.Shortcodes

	s.registerAssets()
	s.registerShortcodes()
	s.registerRoutes()
	s.launchGoroutines()

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
					middleware.WithErrorHandler(http.StatusNotFound, http.HandlerFunc(handlers.NotFound)),
					middleware.WithErrorHandler(http.StatusInternalServerError, http.HandlerFunc(handlers.ServerError)),
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
				metrics.CollectRequestStats(),
				middleware.Compress(),
				middleware.Headers(
					middleware.WithHeader("X-Content-Type-Options", "nosniff"),
					middleware.WithHeader("Content-Security-Policy", "default-src 'self' https://chameth.com/ https://u.c5h.io/ 'nonce-littlefoot-ae805b14'; style-src 'self';"),
					middleware.WithHeader("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload"),
					middleware.WithHeader("Referrer-Policy", "no-referrer-when-downgrade"),
				),
				routing.ApplyRedirects(),
				sudo.Middleware,
				middleware.Recover(),
			),
		)(s.Routes.Public),
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

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Failed to shutdown HTTP server", "error", err)
		panic(err)
	}
}
