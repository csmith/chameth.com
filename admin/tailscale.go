package admin

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"chameth.com/chameth.com/features/routing"
	"github.com/csmith/middleware"
	"os"
	"tailscale.com/tsnet"
)

func RegisterGoroutine(ts *tsnet.Server, rm *routing.Manager) func() {
	return func() {
		if err := start(ts, rm); err != nil {
			slog.Error("Failed to start admin interface", "error", err)
			os.Exit(1)
		}
	}
}

func start(s *tsnet.Server, rm *routing.Manager) error {
	if err := s.Start(); err != nil {
		return err
	}

	httpListener, err := s.Listen("tcp", ":80")
	if err != nil {
		return err
	}

	httpsListener, err := s.ListenTLS("tcp", ":443")
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpsURL := "https://" + fullHostName(s) + r.URL.Path
			if r.URL.RawQuery != "" {
				httpsURL += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
		}),
	}

	httpsServer := &http.Server{
		Handler: middleware.Chain(
			middleware.WithMiddleware(
				middleware.CacheControl(
					middleware.WithCacheTimes(map[string]time.Duration{
						"application/*":   time.Hour * 24 * 365,
						"font/*":          time.Hour * 24 * 365,
						"image/*":         time.Hour * 24 * 365,
						"text/css":        time.Hour * 24 * 365,
						"text/javascript": time.Hour * 24 * 365,
					}),
				),
				middleware.CrossOriginProtection(),
				middleware.Recover(middleware.WithPanicLogger(func(r *http.Request, err any) {
					slog.Error("Panic serving admin site", "url", r.RequestURI, "error", err)
				})),
			),
		)(rm.Admin),
	}

	go httpServer.Serve(httpListener)
	go httpsServer.Serve(httpsListener)

	return nil
}

func fullHostName(s *tsnet.Server) string {
	lc, err := s.LocalClient()
	if err != nil {
		slog.Error("Failed to get local client", "error", err)
		return ""
	}

	status, err := lc.Status(context.Background())
	if err != nil {
		slog.Error("Failed to get local status", "error", err)
		return ""
	}

	fullHostname := status.Self.DNSName
	if len(fullHostname) > 0 && fullHostname[len(fullHostname)-1] == '.' {
		fullHostname = fullHostname[:len(fullHostname)-1]
	}

	return fullHostname
}
