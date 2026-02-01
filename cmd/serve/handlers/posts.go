package handlers

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"chameth.com/chameth.com/cmd/serve/assets"
	"chameth.com/chameth.com/cmd/serve/content"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/syndication"
	"chameth.com/chameth.com/cmd/serve/db"
	"chameth.com/chameth.com/cmd/serve/templates"
	"chameth.com/chameth.com/cmd/serve/templates/includes"
)

func Post(w http.ResponseWriter, r *http.Request) {
	post, err := db.GetPostByPath(r.URL.Path)
	if err != nil {
		slog.Error("Failed to find post by path", "error", err, "path", r.URL.Path)
		ServerError(w, r)
		return
	}

	if post.Path != r.URL.Path {
		http.Redirect(w, r, post.Path, http.StatusPermanentRedirect)
		return
	}

	renderedContent, err := content.RenderContent("post", post.ID, post.Content, post.Path)
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
	ogPath, err := db.GetOpenGraphImageForEntity("post", post.ID)
	if err == nil && ogPath != "" {
		ogImage = fmt.Sprintf("https://chameth.com%s", ogPath)
	}

	relatedPosts, err := content.GetRelatedPosts(post.ID)
	if err != nil {
		slog.Error("Failed to get related posts", "post_id", post.ID, "error", err)
		// Continue without related posts rather than erroring
		relatedPosts = nil
	}

	syndicationInfo, err := syndication.Render(post.Path)
	if err != nil {
		slog.Error("Failed to get syndication markup", "post_id", post.ID, "error", err)
		syndicationInfo = ""
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
			PageData: templates.PageData{
				Title:        fmt.Sprintf("%s · Chameth.com", post.Title),
				Stylesheet:   assets.GetStylesheetPath(),
				CanonicalUrl: fmt.Sprintf("https://chameth.com%s", post.Path),
				OpenGraph: templates.OpenGraphHeaders{
					Image: ogImage,
					Type:  "article",
				},
				RecentPosts: content.RecentPosts(),
			},
			SyndicationInfo: template.HTML(syndicationInfo),
		},
	})
	if err != nil {
		slog.Error("Failed to render post template", "error", err, "path", r.URL.Path)
	}
}

func PostsList(w http.ResponseWriter, r *http.Request) {
	posts, err := db.GetAllPosts()
	if err != nil {
		slog.Error("Failed to get all posts", "error", err)
		ServerError(w, r)
		return
	}

	var postLinks []includes.PostLinkData
	for _, p := range posts {
		postLinks = append(postLinks, content.CreatePostLink(p.Path))
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderPosts(w, templates.PostsData{
		Posts: postLinks,
		PageData: templates.PageData{
			Title:        "Posts · Chameth.com",
			Stylesheet:   assets.GetStylesheetPath(),
			CanonicalUrl: "https://chameth.com/posts/",
			RecentPosts:  content.RecentPosts(),
		},
	})
	if err != nil {
		slog.Error("Failed to render posts template", "error", err)
	}
}
