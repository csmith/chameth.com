package feeds

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/metrics"
	"chameth.com/chameth.com/features/poems"
	"chameth.com/chameth.com/features/posts"
	"chameth.com/chameth.com/features/snippets"
	"chameth.com/chameth.com/templates"
)

func handleAllPosts(w http.ResponseWriter, r *http.Request) {
	renderPostsFeed(w, r, "Chameth.com", "all", 5, templates.SiteURL()+"/index.xml")
}

func handleLongPosts(w http.ResponseWriter, r *http.Request) {
	renderPostsFeed(w, r, "Chameth.com - long posts", "long", 5, templates.SiteURL()+"/long.xml")
}

func handleShortPosts(w http.ResponseWriter, r *http.Request) {
	renderPostsFeed(w, r, "Chameth.com - short posts", "short", 5, templates.SiteURL()+"/short.xml")
}

func handlePoems(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Serving feed", "type", "poems", "useragent", r.UserAgent())
	metrics.RecordFeedRequest("poems", r.UserAgent())

	allPoems, err := poems.GetRecentPoemsWithContent(r.Context(), 5)
	if err != nil {
		slog.Error("Failed to get recent poems for feed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var feedItems []FeedItem
	for _, poem := range allPoems {
		renderedContent, err := content.RenderContent(r.Context(), "poem", poem.ID, poem.Poem, poem.Path)
		if err != nil {
			slog.Error("Failed to render poem content for feed", "poem", poem.Title, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		absoluteContent, err := makeURLsAbsolute(string(renderedContent), templates.SiteURL())
		if err != nil {
			slog.Error("Failed to make URLs absolute for feed", "poem", poem.Title, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		feedItems = append(feedItems, FeedItem{
			Title:   poem.Title,
			Link:    templates.SiteURL() + poem.Path,
			Updated: poem.Date.Format("2006-01-02T15:04:05Z"),
			Content: absoluteContent,
		})
	}

	var lastUpdated string
	if len(allPoems) > 0 {
		lastUpdated = allPoems[0].Date.Format("2006-01-02T15:04:05Z")
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = renderAtom(w, AtomData{
		FeedTitle:       "Chameth.com - poems",
		FeedSelfLink:    templates.SiteURL() + "/poems/feed.xml",
		FeedLastUpdated: lastUpdated,
		FeedItems:       feedItems,
	})
	if err != nil {
		slog.Error("Failed to render atom feed", "error", err)
	}
}

func handleSnippets(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Serving feed", "type", "snippets", "useragent", r.UserAgent())
	metrics.RecordFeedRequest("snippets", r.UserAgent())

	allSnippets, err := snippets.GetRecentSnippetsWithContent(r.Context(), 5)
	if err != nil {
		slog.Error("Failed to get recent snippets for feed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var feedItems []FeedItem
	for _, snippet := range allSnippets {
		renderedContent, err := content.RenderContent(r.Context(), "snippet", snippet.ID, snippet.Content, snippet.Path)
		if err != nil {
			slog.Error("Failed to render snippet content for feed", "snippet", snippet.Title, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		absoluteContent, err := makeURLsAbsolute(string(renderedContent), templates.SiteURL())
		if err != nil {
			slog.Error("Failed to make URLs absolute for feed", "snippet", snippet.Title, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		feedItems = append(feedItems, FeedItem{
			Title:   snippet.Title,
			Link:    templates.SiteURL() + snippet.Path,
			Updated: "1970-01-01T00:00:00Z",
			Content: absoluteContent,
		})
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = renderAtom(w, AtomData{
		FeedTitle:       "Chameth.com - snippets",
		FeedSelfLink:    templates.SiteURL() + "/snippets/feed.xml",
		FeedLastUpdated: "1970-01-01T00:00:00Z",
		FeedItems:       feedItems,
	})
	if err != nil {
		slog.Error("Failed to render atom feed", "error", err)
	}
}

func handleFilmReviews(w http.ResponseWriter, r *http.Request) {
	renderFilmReviewsFeed(w, r, "Chameth.com - film reviews", 5, templates.SiteURL()+"/films/reviews/feed.xml")
}

func renderPostsFeed(w http.ResponseWriter, r *http.Request, title, format string, limit int, selfLink string) {
	slog.Debug("Serving feed", "type", "posts", "format", format, "useragent", r.UserAgent())
	metrics.RecordFeedRequest(format, r.UserAgent())

	var postList []posts.Post
	var err error

	if format == "all" {
		postList, err = posts.GetRecentPostsWithContent(r.Context(), limit)
	} else {
		postList, err = posts.GetRecentPostsWithContentByFormat(r.Context(), limit, format)
	}

	if err != nil {
		slog.Error("Failed to get recent posts for feed", "error", err, "format", format)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	feedItems, err := renderPostItems(r.Context(), postList)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var lastUpdated string
	if len(postList) > 0 {
		lastUpdated = postList[0].Date.Format("2006-01-02T15:04:05Z")
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = renderAtom(w, AtomData{
		FeedTitle:       title,
		FeedSelfLink:    selfLink,
		FeedLastUpdated: lastUpdated,
		FeedItems:       feedItems,
	})
	if err != nil {
		slog.Error("Failed to render atom feed", "error", err)
	}
}

func renderFilmReviewsFeed(w http.ResponseWriter, r *http.Request, title string, limit int, selfLink string) {
	slog.Debug("Serving feed", "type", "filmreviews", "useragent", r.UserAgent())
	metrics.RecordFeedRequest("filmreviews", r.UserAgent())

	reviews, err := films.GetRecentPublishedFilmReviewsWithFilmAndPosters(r.Context(), limit)
	if err != nil {
		slog.Error("Failed to get recent film reviews for feed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var feedItems []FeedItem
	for _, review := range reviews {
		var content strings.Builder
		content.WriteString("<p>")
		fmt.Fprintf(&content, "<strong>Rating:</strong> %d/10", review.FilmReview.Rating)
		if review.FilmReview.IsRewatch {
			content.WriteString(" (Rewatch)")
		}
		content.WriteString("</p>")

		if review.FilmReview.ReviewText != "" {
			fmt.Fprintf(&content, "<p>%s</p>", review.FilmReview.ReviewText)
		}

		reviewURL := templates.SiteURL() + review.Film.Path

		feedItems = append(feedItems, FeedItem{
			Title:   review.Film.Title,
			Link:    reviewURL,
			Updated: review.FilmReview.WatchedDate.Format("2006-01-02T15:04:05Z"),
			Content: content.String(),
		})
	}

	var lastUpdated string
	if len(reviews) > 0 {
		lastUpdated = reviews[0].FilmReview.WatchedDate.Format("2006-01-02T15:04:05Z")
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = renderAtom(w, AtomData{
		FeedTitle:       title,
		FeedSelfLink:    selfLink,
		FeedLastUpdated: lastUpdated,
		FeedItems:       feedItems,
	})
	if err != nil {
		slog.Error("Failed to render atom feed", "error", err)
	}
}

func renderPostItems(ctx context.Context, postList []posts.Post) ([]FeedItem, error) {
	var feedItems []FeedItem
	for _, post := range postList {
		renderedContent, err := content.RenderContent(ctx, "post", post.ID, post.Content, post.Path)
		if err != nil {
			slog.Error("Failed to render post content for feed", "post", post.Title, "error", err)
			return nil, err
		}

		absoluteContent, err := makeURLsAbsolute(string(renderedContent), templates.SiteURL())
		if err != nil {
			slog.Error("Failed to make URLs absolute for feed", "post", post.Title, "error", err)
			return nil, err
		}

		feedItems = append(feedItems, FeedItem{
			Title:   post.Title,
			Link:    templates.SiteURL() + post.Path,
			Updated: post.Date.Format("2006-01-02T15:04:05Z"),
			Content: absoluteContent,
		})
	}

	return feedItems, nil
}

const relatedFeedPrefix = "/feeds/posts/"

// maxFeedSlugs bounds the slug list per category on the public endpoint.
const maxFeedSlugs = 20

var relatedFeedSlugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

func handleRelatedPostsFeed(w http.ResponseWriter, r *http.Request) {
	likes, unlikes, ok := parseRelatedFeedSlugs(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	canonicalPath := relatedFeedPath(likes, unlikes)
	if r.URL.Path != canonicalPath {
		w.Header().Set("Location", canonicalPath)
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}

	slog.Debug("Serving feed", "type", "posts-related", "useragent", r.UserAgent())
	metrics.RecordFeedRequest("posts-related", r.UserAgent())

	postList, err := posts.GetRecentPostsBySimilarity(r.Context(), likes, unlikes, 5)
	if err != nil {
		slog.Error("Failed to get related posts for feed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	feedItems, err := renderPostItems(r.Context(), postList)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var lastUpdated string
	if len(postList) > 0 {
		lastUpdated = postList[0].Date.Format("2006-01-02T15:04:05Z")
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = renderAtom(w, AtomData{
		FeedTitle:       relatedFeedTitle(likes, unlikes),
		FeedSelfLink:    templates.SiteURL() + canonicalPath,
		FeedLastUpdated: lastUpdated,
		FeedItems:       feedItems,
	})
	if err != nil {
		slog.Error("Failed to render atom feed", "error", err)
	}
}

// parseRelatedFeedSlugs parses the like/unlike slug lists out of a
// /feeds/posts/... path. It returns ok=false for anything invalid: unknown or
// repeated categories, odd path elements, or malformed slugs.
func parseRelatedFeedSlugs(path string) (likes, unlikes []string, ok bool) {
	rest := strings.TrimPrefix(path, relatedFeedPrefix)
	if rest == path || rest == "" {
		return nil, nil, false
	}
	rest = strings.TrimSuffix(rest, "/")

	segments := strings.Split(rest, "/")
	if len(segments)%2 != 0 {
		return nil, nil, false
	}

	seen := make(map[string]bool)
	for i := 0; i < len(segments); i += 2 {
		category, list := segments[i], segments[i+1]
		if (category != "like" && category != "unlike") || seen[category] {
			return nil, nil, false
		}
		seen[category] = true

		slugs := strings.Split(list, ",")
		if len(slugs) > maxFeedSlugs {
			return nil, nil, false
		}
		for _, slug := range slugs {
			if !relatedFeedSlugRegex.MatchString(slug) {
				return nil, nil, false
			}
		}

		if category == "like" {
			likes = slugs
		} else {
			unlikes = slugs
		}
	}

	return sortAndDedupe(likes), sortAndDedupe(unlikes), true
}

// sortAndDedupe orders slugs lexicographically and removes duplicates so the
// canonical path is stable regardless of the order slugs were supplied in.
func sortAndDedupe(slugs []string) []string {
	slices.Sort(slugs)
	return slices.Compact(slugs)
}

// relatedFeedPath builds the canonical path: likes first, then unlikes,
// each slug list sorted.
func relatedFeedPath(likes, unlikes []string) string {
	var b strings.Builder
	b.WriteString(relatedFeedPrefix)
	if len(likes) > 0 {
		b.WriteString("like/")
		b.WriteString(strings.Join(likes, ","))
		b.WriteString("/")
	}
	if len(unlikes) > 0 {
		b.WriteString("unlike/")
		b.WriteString(strings.Join(unlikes, ","))
		b.WriteString("/")
	}
	return b.String()
}

func relatedFeedTitle(likes, unlikes []string) string {
	title := "Chameth.com - posts"
	if len(likes) > 0 {
		title += " like " + strings.Join(likes, ", ")
	}
	if len(unlikes) > 0 {
		if len(likes) > 0 {
			title += " but"
		}
		title += " not " + strings.Join(unlikes, ", ")
	}
	return title
}
