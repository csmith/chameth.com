package feeds

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/metrics"
	"chameth.com/chameth.com/features/poems"
	"chameth.com/chameth.com/features/posts"
	"chameth.com/chameth.com/features/snippets"
)

func HandleAllPosts(w http.ResponseWriter, r *http.Request) {
	renderPostsFeed(w, r, "Chameth.com", "all", 5, "https://chameth.com/index.xml")
}

func HandleLongPosts(w http.ResponseWriter, r *http.Request) {
	renderPostsFeed(w, r, "Chameth.com - long posts", "long", 5, "https://chameth.com/long.xml")
}

func HandleShortPosts(w http.ResponseWriter, r *http.Request) {
	renderPostsFeed(w, r, "Chameth.com - short posts", "short", 5, "https://chameth.com/short.xml")
}

func HandlePoems(w http.ResponseWriter, r *http.Request) {
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

		absoluteContent, err := makeURLsAbsolute(string(renderedContent), "https://chameth.com")
		if err != nil {
			slog.Error("Failed to make URLs absolute for feed", "poem", poem.Title, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		feedItems = append(feedItems, FeedItem{
			Title:   poem.Title,
			Link:    fmt.Sprintf("https://chameth.com%s", poem.Path),
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
		FeedSelfLink:    "https://chameth.com/poems/feed.xml",
		FeedLastUpdated: lastUpdated,
		FeedItems:       feedItems,
	})
	if err != nil {
		slog.Error("Failed to render atom feed", "error", err)
	}
}

func HandleSnippets(w http.ResponseWriter, r *http.Request) {
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

		absoluteContent, err := makeURLsAbsolute(string(renderedContent), "https://chameth.com")
		if err != nil {
			slog.Error("Failed to make URLs absolute for feed", "snippet", snippet.Title, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		feedItems = append(feedItems, FeedItem{
			Title:   snippet.Title,
			Link:    fmt.Sprintf("https://chameth.com%s", snippet.Path),
			Updated: "1970-01-01T00:00:00Z",
			Content: absoluteContent,
		})
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = renderAtom(w, AtomData{
		FeedTitle:       "Chameth.com - snippets",
		FeedSelfLink:    "https://chameth.com/snippets/feed.xml",
		FeedLastUpdated: "1970-01-01T00:00:00Z",
		FeedItems:       feedItems,
	})
	if err != nil {
		slog.Error("Failed to render atom feed", "error", err)
	}
}

func HandleFilmReviews(w http.ResponseWriter, r *http.Request) {
	renderFilmReviewsFeed(w, r, "Chameth.com - film reviews", 5, "https://chameth.com/films/reviews/feed.xml")
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

	var feedItems []FeedItem
	for _, post := range postList {
		renderedContent, err := content.RenderContent(r.Context(), "post", post.ID, post.Content, post.Path)
		if err != nil {
			slog.Error("Failed to render post content for feed", "post", post.Title, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		absoluteContent, err := makeURLsAbsolute(string(renderedContent), "https://chameth.com")
		if err != nil {
			slog.Error("Failed to make URLs absolute for feed", "post", post.Title, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		feedItems = append(feedItems, FeedItem{
			Title:   post.Title,
			Link:    fmt.Sprintf("https://chameth.com%s", post.Path),
			Updated: post.Date.Format("2006-01-02T15:04:05Z"),
			Content: absoluteContent,
		})
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

		reviewURL := fmt.Sprintf("https://chameth.com%s", review.Film.Path)

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
