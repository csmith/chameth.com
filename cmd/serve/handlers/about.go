package handlers

import (
	"log/slog"
	"net/http"

	"chameth.com/chameth.com/cmd/serve/assets"
	"chameth.com/chameth.com/cmd/serve/content"
	"chameth.com/chameth.com/cmd/serve/templates"
	"chameth.com/chameth.com/cmd/serve/templates/includes"
)

func About(w http.ResponseWriter, r *http.Request) {
	posts := content.RecentPosts()[:3]

	var links []includes.PostLinkData
	for _, p := range posts {
		links = append(links, content.CreatePostLink(p.Url))
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err := templates.RenderAbout(w, templates.AboutData{
		HighlightedPosts: links,
		Interests: templates.AboutInterests{
			Books: []string{
				"Snow Crash",
				"Dune",
				"A Memory Called Empire",
				"Dungeon Crawler Carl",
			},
			Languages: []string{
				"Go",
				"Kotlin",
				"Android",
			},
			BoardGames: []string{
				"Cartographers",
				"Terraforming Mars",
				"Fluxx",
				"Lovecraft Letter",
			},
			Films: []string{
				"Easy A",
				"Hackers",
				"The Matrix",
				"Everything Everywhere All at Once",
			},
			VideoGames: []string{
				"Warframe",
				"Final Fantasy XIV",
				"Factorio",
				"Rocket League",
			},
		},
		PageData: templates.PageData{
			Title:        "Chameth.com: the personal website of Chris Smith",
			Stylesheet:   assets.GetStylesheetPath(),
			CanonicalUrl: "https://chameth.com/",
			RecentPosts:  content.RecentPosts(),
			OpenGraph: templates.OpenGraphHeaders{
				Type:  "website",
				Image: "/screenshot.png",
			},
		},
	})
	if err != nil {
		slog.Error("Failed to render about template", "error", err)
	}
}
