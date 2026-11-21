package admin

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"chameth.com/chameth.com/features/routing"
	"chameth.com/x/tailscale"
	"github.com/csmith/middleware"
	"tailscale.com/tsnet"
)

func RegisterGoroutine(ctx context.Context, ts *tsnet.Server, rm *routing.Manager) func() {
	return func() {
		if err := serve(ctx, ts, rm); err != nil && ctx.Err() == nil {
			slog.Error("Failed to serve admin interface", "error", err)
			os.Exit(1)
		}
	}
}

func serve(ctx context.Context, s *tsnet.Server, rm *routing.Manager) error {
	handler, err := requireLogin(
		rm.Admin,
		func(ctx context.Context) (string, error) {
			name, err := tailscale.Up(ctx, s)
			if err != nil {
				return "", err
			}
			return "https://" + name, nil
		},
	)
	if err != nil {
		return err
	}

	name, err := tailscale.Up(ctx, s)
	if err != nil {
		return err
	}

	return tailscale.Serve(
		ctx,
		s,
		name,
		middleware.Chain(
			middleware.WithMiddleware(
				middleware.CacheControl(
					middleware.WithCacheTimes(
						map[string]time.Duration{
							"application/*":   time.Hour * 24 * 365,
							"font/*":          time.Hour * 24 * 365,
							"image/*":         time.Hour * 24 * 365,
							"text/css":        time.Hour * 24 * 365,
							"text/javascript": time.Hour * 24 * 365,
						},
					),
				),
				middleware.CrossOriginProtection(),
				middleware.Recover(
					middleware.WithPanicLogger(
						func(r *http.Request, err any) {
							slog.Error("Panic serving admin site", "url", r.RequestURI, "error", err)
						},
					),
				),
			),
		)(handler),
	)
}
