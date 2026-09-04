package char

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/wow"
	"tailscale.com/tsnet"
)

const (
	shortcodeVersion = 1
	refreshFrequency = 4 * time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterData(
		"wowchar",
		shortcodeVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[Data], error) {
			return retrieve(ctx, ts.HTTPClient(), args)
		},
		render,
	)
}

func parseArgs(args []string) (realm, name string, at time.Time, err error) {
	if len(args) < 2 || len(args) > 3 {
		return "", "", time.Time{}, fmt.Errorf("wowchar requires 2 or 3 arguments (realm character [date])")
	}

	realm, name = args[0], args[1]
	if len(args) == 3 {
		at, err = time.Parse("2006-01-02", args[2])
		if err != nil {
			return "", "", time.Time{}, fmt.Errorf("invalid date: %s (expected YYYY-MM-DD)", args[2])
		}
	}

	return realm, name, at, nil
}

func retrieve(ctx context.Context, client *http.Client, args []string) (shortcodes.Result[Data], error) {
	realm, name, at, err := parseArgs(args)
	if err != nil {
		return shortcodes.Result[Data]{}, err
	}

	atParam := ""
	if !at.IsZero() {
		atParam = at.Format("2006-01-02")
	}

	c, err := wow.GetCharacter(ctx, client, realm, name, atParam)
	if err != nil {
		return shortcodes.Result[Data]{}, fmt.Errorf("failed to fetch character: %w", err)
	}
	if c.Profile == nil {
		return shortcodes.Result[Data]{}, fmt.Errorf("character %s-%s has no profile data", name, realm)
	}

	imagePath, err := ensurePortrait(ctx, client, c, at)
	if err != nil {
		return shortcodes.Result[Data]{}, fmt.Errorf("failed to rehost portrait: %w", err)
	}

	return shortcodes.Result[Data]{
		Data:      buildData(c, imagePath),
		RefreshAt: shortcodes.RefreshUntil(refreshFrequency, at),
	}, nil
}

func render(_ []string, data Data, _ *shortcodes.Context) (string, error) {
	return renderTemplate(data)
}
