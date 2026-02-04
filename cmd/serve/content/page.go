package content

import (
	"fmt"

	"chameth.com/chameth.com/cmd/serve/assets"
	"chameth.com/chameth.com/cmd/serve/templates"
)

func CreatePageData(title, path string, ogHeaders templates.OpenGraphHeaders) templates.PageData {
	canonicalUrl := ""
	if path != "" {
		canonicalUrl = fmt.Sprintf("https://chameth.com%s", path)
	}

	return templates.PageData{
		Title:        fmt.Sprintf("%s · Chameth.com", title),
		CanonicalUrl: canonicalUrl,
		OpenGraph:    ogHeaders,
		Stylesheet:   assets.GetStylesheetPath(),
		RecentPosts:  RecentPosts(),
	}
}
