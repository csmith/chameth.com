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
	"chameth.com/chameth.com/features"
	"chameth.com/chameth.com/features/contact"
	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/metrics"
	"chameth.com/chameth.com/features/music"
	"chameth.com/chameth.com/features/posts"
	"chameth.com/chameth.com/features/prints"
	"chameth.com/chameth.com/features/projects"
	"chameth.com/chameth.com/features/snippets"
	"chameth.com/chameth.com/features/sudo"
	"chameth.com/chameth.com/features/syndications"
	"chameth.com/chameth.com/features/wow"
	"chameth.com/chameth.com/handlers"
	"github.com/csmith/envflag/v2"
	"github.com/csmith/middleware"
	"github.com/csmith/slogflags"
	"tailscale.com/tsnet"

	_ "chameth.com/chameth.com/features/boardgames/list"
	_ "chameth.com/chameth.com/features/boardgames/played"
	_ "chameth.com/chameth.com/features/contact/form"
	_ "chameth.com/chameth.com/features/films/list"
	_ "chameth.com/chameth.com/features/films/ratingdistribution"
	_ "chameth.com/chameth.com/features/films/recent"
	_ "chameth.com/chameth.com/features/films/review"
	_ "chameth.com/chameth.com/features/films/reviews"
	_ "chameth.com/chameth.com/features/films/search"
	_ "chameth.com/chameth.com/features/films/watched"
	_ "chameth.com/chameth.com/features/media/audio"
	_ "chameth.com/chameth.com/features/media/figure"
	_ "chameth.com/chameth.com/features/media/video"
	_ "chameth.com/chameth.com/features/music/nowplaying"
	_ "chameth.com/chameth.com/features/music/topalbums"
	_ "chameth.com/chameth.com/features/music/topartists"
	_ "chameth.com/chameth.com/features/posts/link"
	_ "chameth.com/chameth.com/features/posts/recent"
	_ "chameth.com/chameth.com/features/shortcodes/link"
	_ "chameth.com/chameth.com/features/shortcodes/nod"
	_ "chameth.com/chameth.com/features/shortcodes/rating"
	_ "chameth.com/chameth.com/features/shortcodes/sidenote"
	_ "chameth.com/chameth.com/features/shortcodes/update"
	_ "chameth.com/chameth.com/features/shortcodes/warning"
	_ "chameth.com/chameth.com/features/syndications/display"
	_ "chameth.com/chameth.com/features/walks/distance"
	_ "chameth.com/chameth.com/features/walks/list"
	_ "chameth.com/chameth.com/features/walks/speed"
	_ "chameth.com/chameth.com/features/wow/wowchar"
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
	assets.RegisterAssets(assetsManager)
	features.RegisterAssets(assetsManager)

	content.AssetsManager = assetsManager
	content.RecentPostsProvider = posts.Recent

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
		if err := admin.Start(ts); err != nil {
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
	mux.Handle("POST /api/contact", http.HandlerFunc(contact.ContactForm))
	mux.Handle("POST /api/form/contact", http.HandlerFunc(contact.ContactFormPost))
	mux.Handle("POST /api/nod", http.HandlerFunc(handlers.Nod))
	mux.Handle("POST /api/form/nod", http.HandlerFunc(handlers.NodForm))
	mux.Handle("GET /api/films/search", http.HandlerFunc(films.SearchFilmsAPI))
	mux.Handle("GET /assets/stylesheets/", handlers.Stylesheet(assetsManager))
	mux.Handle("GET /assets/scripts/", handlers.Scripts(assetsManager))
	mux.Handle("GET /index.xml", http.HandlerFunc(handlers.FullFeed))
	mux.Handle("GET /short.xml", http.HandlerFunc(handlers.ShortPostsFeed))
	mux.Handle("GET /long.xml", http.HandlerFunc(handlers.LongPostsFeed))
	mux.Handle("GET /poems/feed.xml", http.HandlerFunc(handlers.PoemsFeed))
	mux.Handle("GET /snippets/feed.xml", http.HandlerFunc(handlers.SnippetsFeed))
	mux.Handle("GET /films/reviews/feed.xml", http.HandlerFunc(handlers.FilmReviewsFeed))
	mux.Handle("GET /sitemap.xml", http.HandlerFunc(handlers.XmlSiteMap))
	mux.Handle("GET /posts/{$}", http.HandlerFunc(posts.PostsListHandler))
	mux.Handle("GET /prints/{$}", http.HandlerFunc(prints.PrintsListHandler))
	mux.Handle("GET /projects/{$}", http.HandlerFunc(projects.ProjectsListHandler))
	mux.Handle("GET /sitemap/{$}", http.HandlerFunc(handlers.HtmlSiteMap))
	mux.Handle("GET /snippets/{$}", http.HandlerFunc(snippets.SnippetsListHandler))
	mux.Handle("GET /sudo", http.HandlerFunc(sudo.Handler))
	mux.Handle("/", handlers.Content(handlers.StaticAsset(assets.Static)))

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
