package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/templates"
)

func Post(w http.ResponseWriter, r *http.Request) {
	post, err := db.GetPostByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find post by path", "error", err, "path", r.URL.Path)
		ServerError(w, r)
		return
	}

	if post.Path != r.URL.Path {
		http.Redirect(w, r, post.Path, http.StatusPermanentRedirect)
		return
	}

	renderedContent, err := content.RenderContent(r.Context(), "post", post.ID, post.Content, post.Path)
	if err != nil {
		slog.Error("Failed to render post content", "post", post.Title, "error", err)
		ServerError(w, r)
		return
	}

	yearsOld := int(time.Since(post.Date).Hours() / 24 / 365)
	showWarning := yearsOld >= 5

	summary := post.Content
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}

	var ogImage string
	ogPath, err := db.GetOpenGraphImageForEntity(r.Context(), "post", post.ID)
	if err == nil && ogPath != "" {
		ogImage = fmt.Sprintf("https://chameth.com%s", ogPath)
	}

	relatedPosts, err := content.GetRelatedPosts(r.Context(), post.ID)
	if err != nil {
		slog.Error("Failed to get related posts", "post_id", post.ID, "error", err)
		// Continue without related posts rather than erroring
		relatedPosts = nil
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderPost(w, templates.PostData{
		PostContent: renderedContent,
		PostFormat:  post.Format,
		ArticleData: templates.ArticleData{
			ArticleTitle:   post.Title,
			ArticleSummary: summary,
			ArticleDate: templates.ArticleDate{
				Iso:         post.Date.Format("2006-01-02"),
				Friendly:    post.Date.Format("Jan 2, 2006"),
				ShowWarning: showWarning,
				YearsOld:    yearsOld,
			},
			RelatedPosts: relatedPosts,
			PageData: content.CreatePageData(r.Context(), post.Title, post.Path, templates.OpenGraphHeaders{
				Image: ogImage,
				Type:  "article",
			}),
		},
	})
	if err != nil {
		slog.Error("Failed to render post template", "error", err, "path", r.URL.Path)
	}
}

func PostsList(w http.ResponseWriter, r *http.Request) {
	posts, err := db.GetAllPosts(r.Context())
	if err != nil {
		slog.Error("Failed to get all posts", "error", err)
		ServerError(w, r)
		return
	}

	var postPaths []string
	for _, p := range posts {
		postPaths = append(postPaths, p.Path)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderPosts(w, templates.PostsData{
		Posts:    postPaths,
		PageData: content.CreatePageData(r.Context(), "Posts", "/posts/", templates.OpenGraphHeaders{}),
	})
	if err != nil {
		slog.Error("Failed to render posts template", "error", err)
	}
}
