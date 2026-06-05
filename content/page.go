package content

import (
	"context"
	"fmt"
	"log/slog"

	"chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/sudo"
	"chameth.com/chameth.com/templates"
)

var RecentPostsProvider func() []templates.RecentPost
var LinksProvider func(context.Context, string) ([]templates.Link, error)
var AssetsManager *assets.Manager
var ShortcodesManager *shortcodes.Manager

func CreatePageData(ctx context.Context, title, path string, ogHeaders templates.OpenGraphHeaders) templates.PageData {
	canonicalUrl := ""
	if path != "" {
		canonicalUrl = fmt.Sprintf("https://chameth.com%s", path)
	}

	links, err := LinksProvider(ctx, path)
	if err != nil {
		slog.Warn("Failed to get links", "path", path, "error", err)
	}

	return templates.PageData{
		Title:        fmt.Sprintf("%s · Chameth.com", title),
		CanonicalUrl: canonicalUrl,
		OpenGraph:    ogHeaders,
		Scripts:      scriptPath(),
		Stylesheet:   stylesheetPath(),
		RecentPosts:  RecentPostsProvider(),
		Component:    ShortcodesManager.NewComponentFunc(&shortcodes.Context{Context: ctx, URL: path}),
		Admin:        sudo.IsAdmin(ctx),
		Links:        links,
	}
}

func stylesheetPath() string {
	_, checksum := AssetsManager.Bundle(assets.PublicCSS)
	return checksum + ".css"
}

func scriptPath() string {
	_, checksum := AssetsManager.Bundle(assets.PublicJS)
	return checksum + ".js"
}
