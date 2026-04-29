package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/metrics"
	"chameth.com/chameth.com/features/snippets"
	"chameth.com/chameth.com/templates"
	"golang.org/x/net/html"
)

func FullFeed(w http.ResponseWriter, r *http.Request) {
	renderFeed(w, r, "Chameth.com", "all", 5, "https://chameth.com/index.xml")
}

func LongPostsFeed(w http.ResponseWriter, r *http.Request) {
	renderFeed(w, r, "Chameth.com - long posts", "long", 5, "https://chameth.com/long.xml")
}

func ShortPostsFeed(w http.ResponseWriter, r *http.Request) {
	renderFeed(w, r, "Chameth.com - short posts", "short", 5, "https://chameth.com/short.xml")
}

func PoemsFeed(w http.ResponseWriter, r *http.Request) {
	renderPoemsFeed(w, r, "Chameth.com - poems", 5, "https://chameth.com/poems/feed.xml")
}

func SnippetsFeed(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Serving feed", "type", "snippets", "useragent", r.UserAgent())
	metrics.RecordFeedRequest("snippets", r.UserAgent())

	allSnippets, err := snippets.GetRecentSnippetsWithContent(r.Context(), 5)
	if err != nil {
		slog.Error("Failed to get recent snippets for feed", "error", err)
		ServerError(w, r)
		return
	}

	var feedItems []templates.FeedItem
	for _, snippet := range allSnippets {
		renderedContent, err := content.RenderContent(r.Context(), "snippet", snippet.ID, snippet.Content, snippet.Path)
		if err != nil {
			slog.Error("Failed to render snippet content for feed", "snippet", snippet.Title, "error", err)
			ServerError(w, r)
			return
		}

		absoluteContent, err := MakeURLsAbsolute(string(renderedContent), "https://chameth.com")
		if err != nil {
			slog.Error("Failed to make URLs absolute for feed", "snippet", snippet.Title, "error", err)
			ServerError(w, r)
			return
		}

		feedItems = append(feedItems, templates.FeedItem{
			Title:   snippet.Title,
			Link:    fmt.Sprintf("https://chameth.com%s", snippet.Path),
			Updated: "1970-01-01T00:00:00Z",
			Content: absoluteContent,
		})
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderAtom(w, templates.AtomData{
		FeedTitle:       "Chameth.com - snippets",
		FeedSelfLink:    "https://chameth.com/snippets/feed.xml",
		FeedLastUpdated: "1970-01-01T00:00:00Z",
		FeedItems:       feedItems,
	})
	if err != nil {
		slog.Error("Failed to render atom feed", "error", err)
	}
}

func FilmReviewsFeed(w http.ResponseWriter, r *http.Request) {
	renderFilmReviewsFeed(w, r, "Chameth.com - film reviews", 5, "https://chameth.com/films/reviews/feed.xml")
}

func renderFeed(w http.ResponseWriter, r *http.Request, title, format string, limit int, selfLink string) {
	slog.Debug("Serving feed", "type", "posts", "format", format, "useragent", r.UserAgent())
	metrics.RecordFeedRequest(format, r.UserAgent())

	var posts []db.Post
	var err error

	if format == "all" {
		posts, err = db.GetRecentPostsWithContent(r.Context(), limit)
	} else {
		posts, err = db.GetRecentPostsWithContentByFormat(r.Context(), limit, format)
	}

	if err != nil {
		slog.Error("Failed to get recent posts for feed", "error", err, "format", format)
		ServerError(w, r)
		return
	}

	var feedItems []templates.FeedItem
	for _, post := range posts {
		renderedContent, err := content.RenderContent(r.Context(), "post", post.ID, post.Content, post.Path)
		if err != nil {
			slog.Error("Failed to render post content for feed", "post", post.Title, "error", err)
			ServerError(w, r)
			return
		}

		absoluteContent, err := MakeURLsAbsolute(string(renderedContent), "https://chameth.com")
		if err != nil {
			slog.Error("Failed to make URLs absolute for feed", "post", post.Title, "error", err)
			ServerError(w, r)
			return
		}

		feedItems = append(feedItems, templates.FeedItem{
			Title:   post.Title,
			Link:    fmt.Sprintf("https://chameth.com%s", post.Path),
			Updated: post.Date.Format("2006-01-02T15:04:05Z"),
			Content: absoluteContent,
		})
	}

	var lastUpdated string
	if len(posts) > 0 {
		lastUpdated = posts[0].Date.Format("2006-01-02T15:04:05Z")
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderAtom(w, templates.AtomData{
		FeedTitle:       title,
		FeedSelfLink:    selfLink,
		FeedLastUpdated: lastUpdated,
		FeedItems:       feedItems,
	})
	if err != nil {
		slog.Error("Failed to render atom feed", "error", err)
	}
}

func renderPoemsFeed(w http.ResponseWriter, r *http.Request, title string, limit int, selfLink string) {
	slog.Debug("Serving feed", "type", "poems", "useragent", r.UserAgent())
	metrics.RecordFeedRequest("poems", r.UserAgent())

	poems, err := db.GetRecentPoemsWithContent(r.Context(), limit)
	if err != nil {
		slog.Error("Failed to get recent poems for feed", "error", err)
		ServerError(w, r)
		return
	}

	var feedItems []templates.FeedItem
	for _, poem := range poems {
		renderedContent, err := content.RenderContent(r.Context(), "poem", poem.ID, poem.Poem, poem.Path)
		if err != nil {
			slog.Error("Failed to render poem content for feed", "poem", poem.Title, "error", err)
			ServerError(w, r)
			return
		}

		absoluteContent, err := MakeURLsAbsolute(string(renderedContent), "https://chameth.com")
		if err != nil {
			slog.Error("Failed to make URLs absolute for feed", "poem", poem.Title, "error", err)
			ServerError(w, r)
			return
		}

		feedItems = append(feedItems, templates.FeedItem{
			Title:   poem.Title,
			Link:    fmt.Sprintf("https://chameth.com%s", poem.Path),
			Updated: poem.Date.Format("2006-01-02T15:04:05Z"),
			Content: absoluteContent,
		})
	}

	var lastUpdated string
	if len(poems) > 0 {
		lastUpdated = poems[0].Date.Format("2006-01-02T15:04:05Z")
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderAtom(w, templates.AtomData{
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
		ServerError(w, r)
		return
	}

	var feedItems []templates.FeedItem
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

		feedItems = append(feedItems, templates.FeedItem{
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
	err = templates.RenderAtom(w, templates.AtomData{
		FeedTitle:       title,
		FeedSelfLink:    selfLink,
		FeedLastUpdated: lastUpdated,
		FeedItems:       feedItems,
	})
	if err != nil {
		slog.Error("Failed to render atom feed", "error", err)
	}
}

func MakeURLsAbsolute(htmlContent, baseURL string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var processNode func(*html.Node)
	processNode = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i, attr := range n.Attr {
				if (attr.Key == "href" || attr.Key == "src") && strings.HasPrefix(attr.Val, "/") && !strings.HasPrefix(attr.Val, "//") {
					n.Attr[i].Val = baseURL + attr.Val
				}
				if attr.Key == "srcset" {
					n.Attr[i].Val = makeSrcsetAbsolute(attr.Val, baseURL)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processNode(c)
		}
	}

	processNode(doc)

	var buf strings.Builder
	if err := html.Render(&buf, doc); err != nil {
		return "", fmt.Errorf("failed to render HTML: %w", err)
	}

	result := buf.String()
	result = strings.TrimPrefix(result, "<html><head></head><body>")
	result = strings.TrimSuffix(result, "</body></html>")

	return result, nil
}

func makeSrcsetAbsolute(srcset, baseURL string) string {
	parts := strings.Split(srcset, ",")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		urlAndDescriptor := strings.Fields(part)
		if len(urlAndDescriptor) > 0 && strings.HasPrefix(urlAndDescriptor[0], "/") && !strings.HasPrefix(urlAndDescriptor[0], "//") {
			urlAndDescriptor[0] = baseURL + urlAndDescriptor[0]
			parts[i] = strings.Join(urlAndDescriptor, " ")
		}
	}
	return strings.Join(parts, ", ")
}
