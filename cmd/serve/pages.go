package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/csmith/chameth.com/cmd/serve/templates"
)

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	err := templates.RenderNotFound(w, templates.NotFoundData{
		PageData: templates.PageData{
			Title:       "Not found · Chameth.com",
			Stylesheet:  compiledSheetPath,
			RecentPosts: recentPosts,
		},
	})
	if err != nil {
		slog.Error("Failed to render not found template", "error", err)
	}
}

func handleServerError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	err := templates.RenderServerError(w, templates.ServerErrorData{
		PageData: templates.PageData{
			Title:       "Server error · Chameth.com",
			Stylesheet:  compiledSheetPath,
			RecentPosts: recentPosts,
		},
	})
	if err != nil {
		slog.Error("Failed to render not found template", "error", err)
	}
}

func handlePGP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err := templates.RenderPGP(w, templates.PGPData{
		PageData: templates.PageData{
			Title:        "PGP information · Chameth.com",
			CanonicalUrl: "https://chameth.com/pgp/",
			Stylesheet:   compiledSheetPath,
			RecentPosts:  recentPosts,
		},
	})
	if err != nil {
		slog.Error("Failed to render pgp template", "error", err)
	}
}

func handleContent(w http.ResponseWriter, r *http.Request) {
	contentType, err := findContentBySlug(r.URL.Path)
	if err != nil {
		slog.Error("Failed to find content by slug", "error", err, "path", r.URL.Path)
		handleServerError(w, r)
		return
	}

	switch contentType {
	case "poem":
		handlePoem(w, r)
	case "snippet":
		handleSnippet(w, r)
	default:
		// In the future this will be a 404, but for now fall back to 11ty rendered content
		http.FileServer(http.Dir(*files)).ServeHTTP(w, r)
	}
}

func handlePoem(w http.ResponseWriter, r *http.Request) {
	poem, err := getPoemBySlug(r.URL.Path)
	if err != nil {
		slog.Error("Failed to find poem by slug", "error", err, "path", r.URL.Path)
		handleServerError(w, r)
		return
	}

	if poem.Slug != r.URL.Path {
		http.Redirect(w, r, poem.Slug, http.StatusPermanentRedirect)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderPoem(w, templates.PoemData{
		Poem:     poem.Poem,
		Comments: poem.Notes,
		ArticleData: templates.ArticleData{
			ArticleTitle:   poem.Title,
			ArticleSummary: poem.Poem,
			ArticleDate: templates.ArticleDate{
				Iso:         poem.Published.Format("2006-01-02"),
				Friendly:    poem.Published.Format("Jan 2, 2006"),
				ShowWarning: false,
			},
			PageData: templates.PageData{
				Title:        fmt.Sprintf("%s · Chameth.com", poem.Title),
				Stylesheet:   compiledSheetPath,
				CanonicalUrl: fmt.Sprintf("https://chameth.com%s", poem.Slug),
				RecentPosts:  recentPosts,
			},
		},
	})
	if err != nil {
		slog.Error("Failed to render poem template", "error", err, "path", r.URL.Path)
	}
}

func handleSnippet(w http.ResponseWriter, r *http.Request) {
	snippet, err := getSnippetBySlug(r.URL.Path)
	if err != nil {
		slog.Error("Failed to find snippet by slug", "error", err, "path", r.URL.Path)
		handleServerError(w, r)
		return
	}

	if snippet.Slug != r.URL.Path {
		http.Redirect(w, r, snippet.Slug, http.StatusPermanentRedirect)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderSnippet(w, templates.SnippetData{
		SnippetTitle:   snippet.Title,
		SnippetGroup:   snippet.Topic,
		SnippetContent: snippet.Content,
		PageData: templates.PageData{
			Title:        fmt.Sprintf("%s · Chameth.com", snippet.Title),
			Stylesheet:   compiledSheetPath,
			CanonicalUrl: fmt.Sprintf("https://chameth.com%s", snippet.Slug),
			RecentPosts:  recentPosts,
		},
	})
	if err != nil {
		slog.Error("Failed to render snippet template", "error", err, "path", r.URL.Path)
	}
}
