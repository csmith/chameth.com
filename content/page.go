package content

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/sudo"
	"chameth.com/chameth.com/templates"
)

var RecentPostsProvider func() []templates.RecentPost
var AssetsManager *assets.Manager

func CreatePageData(ctx context.Context, title, path string, ogHeaders templates.OpenGraphHeaders) templates.PageData {
	canonicalUrl := ""
	if path != "" {
		canonicalUrl = fmt.Sprintf("https://chameth.com%s", path)
	}

	return templates.PageData{
		Title:        fmt.Sprintf("%s · Chameth.com", title),
		CanonicalUrl: canonicalUrl,
		OpenGraph:    ogHeaders,
		Scripts:      scriptPath(),
		Stylesheet:   stylesheetPath(),
		RecentPosts:  RecentPostsProvider(),
		Component:    shortcodes.NewComponentFunc(&shortcodes.Context{Context: ctx, URL: path}),
		Admin:        sudo.IsAdmin(ctx),
	}
}

func stylesheetPath() string {
	_, checksum := AssetsManager.Bundle(assets.PublicCSS)
	return checksum + ".css"
}

func scriptPath() string {
	_, checksum := AssetsManager.Bundle(assets.PublicCSS)
	return checksum + ".js"
}
