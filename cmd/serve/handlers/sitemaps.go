package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"chameth.com/chameth.com/cmd/serve/assets"
	"chameth.com/chameth.com/cmd/serve/content"
	"chameth.com/chameth.com/cmd/serve/db"
	"chameth.com/chameth.com/cmd/serve/templates"
)

func HtmlSiteMap(w http.ResponseWriter, r *http.Request) {
	poems, err := db.GetAllPoems()
	if err != nil {
		slog.Error("Failed to get all poems", "error", err)
		ServerError(w, r)
		return
	}

	var poemDetails []templates.ContentDetails
	for _, p := range poems {
		poemDetails = append(poemDetails, templates.ContentDetails{
			Title: p.Title,
			Path:  p.Path,
			Date: templates.ContentDate{
				Iso:      p.Date.Format("2006-01-02"),
				Friendly: p.Date.Format("Jan 2, 2006"),
			},
		})
	}

	snippets, err := db.GetAllSnippets()
	if err != nil {
		slog.Error("Failed to get all snippets", "error", err)
		ServerError(w, r)
		return
	}

	var snippetDetails []templates.SnippetDetails
	for _, s := range snippets {
		snippetDetails = append(snippetDetails, templates.SnippetDetails{
			Path: s.Path,
			Name: fmt.Sprintf("%s ➔ %s", s.Topic, s.Title),
		})
	}

	posts, err := db.GetAllPosts()
	if err != nil {
		slog.Error("Failed to get all posts", "error", err)
		ServerError(w, r)
		return
	}

	var postDetails []templates.ContentDetails
	for _, p := range posts {
		postDetails = append(postDetails, templates.ContentDetails{
			Title: p.Title,
			Path:  p.Path,
			Date: templates.ContentDate{
				Iso:      p.Date.Format("2006-01-02"),
				Friendly: p.Date.Format("Jan 2, 2006"),
			},
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderHtmlSiteMap(w, templates.SiteMapData{
		Posts:    postDetails,
		Poems:    poemDetails,
		Snippets: snippetDetails,
		PageData: templates.PageData{
			Title:        "Sitemap · Chameth.com",
			Stylesheet:   assets.GetStylesheetPath(),
			CanonicalUrl: "https://chameth.com/sitemap/",
			RecentPosts:  content.RecentPosts(),
		},
	})
	if err != nil {
		slog.Error("Failed to render site map template", "error", err)
	}
}

func XmlSiteMap(w http.ResponseWriter, r *http.Request) {
	poems, err := db.GetAllPoems()
	if err != nil {
		slog.Error("Failed to get all poems", "error", err)
		ServerError(w, r)
		return
	}

	var poemDetails []templates.ContentDetails
	for _, p := range poems {
		poemDetails = append(poemDetails, templates.ContentDetails{
			Title: p.Title,
			Path:  p.Path,
			Date: templates.ContentDate{
				Iso:      p.Date.Format("2006-01-02"),
				Friendly: p.Date.Format("Jan 2, 2006"),
			},
		})
	}

	snippets, err := db.GetAllSnippets()
	if err != nil {
		slog.Error("Failed to get all snippets", "error", err)
		ServerError(w, r)
		return
	}

	var snippetDetails []templates.SnippetDetails
	for _, s := range snippets {
		snippetDetails = append(snippetDetails, templates.SnippetDetails{
			Path: s.Path,
			Name: fmt.Sprintf("%s ➔ %s", s.Topic, s.Title),
		})
	}

	posts, err := db.GetAllPosts()
	if err != nil {
		slog.Error("Failed to get all posts", "error", err)
		ServerError(w, r)
		return
	}

	var postDetails []templates.ContentDetails
	for _, p := range posts {
		postDetails = append(postDetails, templates.ContentDetails{
			Title: p.Title,
			Path:  p.Path,
			Date: templates.ContentDate{
				Iso:      p.Date.Format("2006-01-02"),
				Friendly: p.Date.Format("Jan 2, 2006"),
			},
		})
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderXmlSiteMap(w, templates.SiteMapData{
		Posts:    postDetails,
		Poems:    poemDetails,
		Snippets: snippetDetails,
	})
	if err != nil {
		slog.Error("Failed to render site map template", "error", err)
	}
}
