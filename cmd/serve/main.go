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

	"chameth.com/chameth.com/admin"
	"chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/features/contact"
	"chameth.com/chameth.com/features/feeds"
	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/metrics"
	"chameth.com/chameth.com/features/music"
	"chameth.com/chameth.com/features/nod"
	"chameth.com/chameth.com/features/posts"
	"chameth.com/chameth.com/features/prints"
	"chameth.com/chameth.com/features/projects"
	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/sitemap"
	"chameth.com/chameth.com/features/snippets"
	"chameth.com/chameth.com/features/sudo"
	"chameth.com/chameth.com/features/syndications"
	"chameth.com/chameth.com/features/wow"
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

	metrics.StartMetricsServer()
	if err := db.Init(metrics.LogQuery); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	assetsManager := assets.NewManager()
	shortcodesManager := shortcodes.NewManager()

	content.AssetsManager = assetsManager
	content.RecentPostsProvider = posts.Recent
	content.ShortcodesManager = shortcodesManager

	registerAssets(assetsManager)
	registerShortcodes(shortcodesManager)

	go posts.UpdateAllPosts(context.Background())
	go syndications.SyndicateAllPosts(context.Background())

	ts := &tsnet.Server{
		Hostname: *tailscaleHost,
		Dir:      *tailscaleDir,
		UserLogf: func(s string, v ...any) {
			slog.Info(fmt.Sprintf(s, v...), "source", "tailscale")
		},
		Logf: func(s string, v ...any) {
			slog.Debug(fmt.Sprintf(s, v...), "source", "tailscale")
		},
	}

	go func() {
		if err := admin.Start(ts, assetsManager); err != nil {
			slog.Error("Failed to start admin interface", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		ts.Up(context.Background())
		music.RunImport(context.Background(), ts.HTTPClient())
	}()

	go wow.RunSync(context.Background())

	mux := http.NewServeMux()
	mux.Handle("POST /api/contact", http.HandlerFunc(contact.HandleJSON))
	mux.Handle("POST /api/form/contact", http.HandlerFunc(contact.HandleForm))
	mux.Handle("POST /api/nod", http.HandlerFunc(nod.HandleJSON))
	mux.Handle("POST /api/form/nod", http.HandlerFunc(nod.HandleForm))
	mux.Handle("GET /api/films/search", http.HandlerFunc(films.HandleSearch))
	mux.Handle("GET /assets/stylesheets/", handlers.Stylesheet(assetsManager))
	mux.Handle("GET /assets/scripts/", handlers.Scripts(assetsManager))
	mux.Handle("GET /index.xml", http.HandlerFunc(feeds.HandleAllPosts))
	mux.Handle("GET /short.xml", http.HandlerFunc(feeds.HandleShortPosts))
	mux.Handle("GET /long.xml", http.HandlerFunc(feeds.HandleLongPosts))
	mux.Handle("GET /poems/feed.xml", http.HandlerFunc(feeds.HandlePoems))
	mux.Handle("GET /snippets/feed.xml", http.HandlerFunc(feeds.HandleSnippets))
	mux.Handle("GET /films/reviews/feed.xml", http.HandlerFunc(feeds.HandleFilmReviews))
	mux.Handle("GET /sitemap.xml", http.HandlerFunc(sitemap.HandleXml))
	mux.Handle("GET /posts/{$}", http.HandlerFunc(posts.HandleList))
	mux.Handle("GET /prints/{$}", http.HandlerFunc(prints.HandleList))
	mux.Handle("GET /projects/{$}", http.HandlerFunc(projects.HandleList))
	mux.Handle("GET /sitemap/{$}", http.HandlerFunc(sitemap.HandleHtml))
	mux.Handle("GET /snippets/{$}", http.HandlerFunc(snippets.HandleList))
	mux.Handle("GET /sudo", http.HandlerFunc(sudo.Handle))
	mux.Handle("/", handlers.Content(handlers.StaticAsset(assetsManager)))

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
				applyRedirects(),
				sudo.Middleware,
				middleware.Recover(),
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
