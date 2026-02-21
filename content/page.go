package content

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/content/shortcodes"
	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/templates"
)

func CreatePageData(ctx context.Context, title, path string, ogHeaders templates.OpenGraphHeaders) templates.PageData {
	canonicalUrl := ""
	if path != "" {
		canonicalUrl = fmt.Sprintf("https://chameth.com%s", path)
	}

	return templates.PageData{
		Title:        fmt.Sprintf("%s · Chameth.com", title),
		CanonicalUrl: canonicalUrl,
		OpenGraph:    ogHeaders,
		Scripts:      assets.GetScriptPath(),
		Stylesheet:   assets.GetStylesheetPath(),
		RecentPosts:  RecentPosts(),
		Component:    shortcodes.NewComponentFunc(&common.Context{Context: ctx, URL: path}),
	}
}
