package posts

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/features/media"
	posttemplates "chameth.com/chameth.com/features/posts/templates"
	parenttemplates "chameth.com/chameth.com/templates"
)

func PostHandler(w http.ResponseWriter, r *http.Request) {
	post, err := GetPostByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find post by path", "error", err, "path", r.URL.Path)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if post.Path != r.URL.Path {
		http.Redirect(w, r, post.Path, http.StatusPermanentRedirect)
		return
	}

	renderedContent, err := content.RenderContent(r.Context(), "post", post.ID, post.Content, post.Path)
	if err != nil {
		slog.Error("Failed to render post content", "post", post.Title, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	yearsOld := int(time.Since(post.Date).Hours() / 24 / 365)
	showWarning := yearsOld >= 5

	summary := post.Content
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}

	var ogImage string
	ogPath, err := media.GetOpenGraphImageForEntity(r.Context(), "post", post.ID)
	if err == nil && ogPath != "" {
		ogImage = parenttemplates.SiteURL() + ogPath
	}

	relatedPosts, err := RelatedPosts(r.Context(), post.ID)
	if err != nil {
		slog.Error("Failed to get related posts", "post_id", post.ID, "error", err)
		relatedPosts = nil
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = posttemplates.RenderPost(w, posttemplates.PostData{
		PostContent:    renderedContent,
		PostFormat:     post.Format,
		ArticleTitle:   post.Title,
		ArticleSummary: summary,
		ArticleDate: parenttemplates.ArticleDate{
			Iso:         post.Date.Format("2006-01-02"),
			Friendly:    post.Date.Format("Jan 2, 2006"),
			ShowWarning: showWarning,
			YearsOld:    yearsOld,
		},
		RelatedPosts:   relatedPosts,
		FeedBuilderURL: feedBuilderURL(post.Path),
		EditLink:       fmt.Sprintf("%s/posts/edit/%d", parenttemplates.AdminURL(), post.ID),
		PageData: content.CreatePageData(r.Context(), post.Title, post.Path, parenttemplates.OpenGraphHeaders{
			Image: ogImage,
			Type:  "article",
		}),
	})
	if err != nil {
		slog.Error("Failed to render post template", "error", err, "path", r.URL.Path)
	}
}

// feedBuilderURL returns the URL for a feed of posts similar to this one, or
// an empty string when the post path cannot be expressed as a feed slug.
func feedBuilderURL(path string) string {
	slug := strings.Trim(path, "/")
	if !feedSlugRegex.MatchString(slug) {
		return ""
	}
	return fmt.Sprintf("/feeds/posts/build/like/%s/", slug)
}

var feedSlugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

func handleList(w http.ResponseWriter, r *http.Request) {
	allPosts, err := GetAllPosts(r.Context())
	if err != nil {
		slog.Error("Failed to get all posts", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var postPaths []string
	for _, p := range allPosts {
		postPaths = append(postPaths, p.Path)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = posttemplates.RenderPosts(w, posttemplates.PostsData{
		Posts:    postPaths,
		PageData: content.CreatePageData(r.Context(), "Posts", "/posts/", parenttemplates.OpenGraphHeaders{}),
	})
	if err != nil {
		slog.Error("Failed to render posts template", "error", err)
	}
}
