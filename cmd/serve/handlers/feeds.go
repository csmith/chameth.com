package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content"
	"chameth.com/chameth.com/cmd/serve/db"
	"chameth.com/chameth.com/cmd/serve/templates"
	"golang.org/x/net/html"
)

func FullFeed(w http.ResponseWriter, r *http.Request) {
	renderFeed(w, r, "Chameth.com", "", 5)
}

func LongPostsFeed(w http.ResponseWriter, r *http.Request) {
	renderFeed(w, r, "Chameth.com - long posts", "long", 5)
}

func ShortPostsFeed(w http.ResponseWriter, r *http.Request) {
	renderFeed(w, r, "Chameth.com - short posts", "short", 5)
}

func PoemsFeed(w http.ResponseWriter, r *http.Request) {
	renderPoemsFeed(w, r, "Chameth.com - poems", 5)
}

func SnippetsFeed(w http.ResponseWriter, r *http.Request) {
	renderSnippetsFeed(w, r, "Chameth.com - snippets", 5)
}

func FilmReviewsFeed(w http.ResponseWriter, r *http.Request) {
	renderFilmReviewsFeed(w, r, "Chameth.com - film reviews", 5)
}

func renderFeed(w http.ResponseWriter, r *http.Request, title, format string, limit int) {
	var posts []db.Post
	var err error

	if format == "" {
		posts, err = db.GetRecentPostsWithContent(limit)
	} else {
		posts, err = db.GetRecentPostsWithContentByFormat(limit, format)
	}

	if err != nil {
		slog.Error("Failed to get recent posts for feed", "error", err, "format", format)
		ServerError(w, r)
		return
	}

	var feedItems []templates.FeedItem
	for _, post := range posts {
		// Render content (shortcodes + markdown)
		renderedContent, err := content.RenderContent("post", post.ID, post.Content)
		if err != nil {
			slog.Error("Failed to render post content for feed", "post", post.Title, "error", err)
			ServerError(w, r)
			return
		}

		// Convert relative URLs to absolute
		absoluteContent, err := makeURLsAbsolute(string(renderedContent), "https://chameth.com")
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

	// Get the last updated date from the most recent post
	var lastUpdated string
	if len(posts) > 0 {
		lastUpdated = posts[0].Date.Format("2006-01-02T15:04:05Z")
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderAtom(w, templates.AtomData{
		FeedTitle:       title,
		FeedLastUpdated: lastUpdated,
		FeedItems:       feedItems,
	})
	if err != nil {
		slog.Error("Failed to render atom feed", "error", err)
	}
}

func renderPoemsFeed(w http.ResponseWriter, r *http.Request, title string, limit int) {
	poems, err := db.GetRecentPoemsWithContent(limit)
	if err != nil {
		slog.Error("Failed to get recent poems for feed", "error", err)
		ServerError(w, r)
		return
	}

	var feedItems []templates.FeedItem
	for _, poem := range poems {
		// Render content (shortcodes + markdown)
		renderedContent, err := content.RenderContent("poem", poem.ID, poem.Poem)
		if err != nil {
			slog.Error("Failed to render poem content for feed", "poem", poem.Title, "error", err)
			ServerError(w, r)
			return
		}

		// Convert relative URLs to absolute
		absoluteContent, err := makeURLsAbsolute(string(renderedContent), "https://chameth.com")
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

	// Get the last updated date from the most recent poem
	var lastUpdated string
	if len(poems) > 0 {
		lastUpdated = poems[0].Date.Format("2006-01-02T15:04:05Z")
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderAtom(w, templates.AtomData{
		FeedTitle:       title,
		FeedLastUpdated: lastUpdated,
		FeedItems:       feedItems,
	})
	if err != nil {
		slog.Error("Failed to render atom feed", "error", err)
	}
}

func renderSnippetsFeed(w http.ResponseWriter, r *http.Request, title string, limit int) {
	snippets, err := db.GetRecentSnippetsWithContent(limit)
	if err != nil {
		slog.Error("Failed to get recent snippets for feed", "error", err)
		ServerError(w, r)
		return
	}

	var feedItems []templates.FeedItem
	for _, snippet := range snippets {
		// Render content (shortcodes + markdown)
		renderedContent, err := content.RenderContent("snippet", snippet.ID, snippet.Content)
		if err != nil {
			slog.Error("Failed to render snippet content for feed", "snippet", snippet.Title, "error", err)
			ServerError(w, r)
			return
		}

		// Convert relative URLs to absolute
		absoluteContent, err := makeURLsAbsolute(string(renderedContent), "https://chameth.com")
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
		FeedTitle:       title,
		FeedLastUpdated: "1970-01-01T00:00:00Z",
		FeedItems:       feedItems,
	})
	if err != nil {
		slog.Error("Failed to render atom feed", "error", err)
	}
}

func renderFilmReviewsFeed(w http.ResponseWriter, r *http.Request, title string, limit int) {
	reviews, err := db.GetRecentPublishedFilmReviewsWithFilmAndPosters(limit)
	if err != nil {
		slog.Error("Failed to get recent film reviews for feed", "error", err)
		ServerError(w, r)
		return
	}

	var feedItems []templates.FeedItem
	for _, review := range reviews {
		var content strings.Builder
		content.WriteString("<p>")
		content.WriteString(fmt.Sprintf("<strong>Rating:</strong> %d/10", review.FilmReview.Rating))
		if review.FilmReview.IsRewatch {
			content.WriteString(" (Rewatch)")
		}
		content.WriteString("</p>")

		if review.FilmReview.ReviewText != "" {
			content.WriteString(fmt.Sprintf("<p>%s</p>", review.FilmReview.ReviewText))
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
		FeedLastUpdated: lastUpdated,
		FeedItems:       feedItems,
	})
	if err != nil {
		slog.Error("Failed to render atom feed", "error", err)
	}
}

// makeURLsAbsolute converts relative URLs in HTML to absolute URLs
func makeURLsAbsolute(htmlContent, baseURL string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var processNode func(*html.Node)
	processNode = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i, attr := range n.Attr {
				// Convert relative URLs to absolute for common attributes
				if (attr.Key == "href" || attr.Key == "src") && strings.HasPrefix(attr.Val, "/") && !strings.HasPrefix(attr.Val, "//") {
					n.Attr[i].Val = baseURL + attr.Val
				}
				// Also handle srcset attributes
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

	// Remove the <html><head></head><body> wrapper that html.Parse adds
	result := buf.String()
	result = strings.TrimPrefix(result, "<html><head></head><body>")
	result = strings.TrimSuffix(result, "</body></html>")

	return result, nil
}

// makeSrcsetAbsolute converts relative URLs in srcset attributes to absolute
func makeSrcsetAbsolute(srcset, baseURL string) string {
	parts := strings.Split(srcset, ",")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		// Split on space to separate URL from descriptor
		urlAndDescriptor := strings.Fields(part)
		if len(urlAndDescriptor) > 0 && strings.HasPrefix(urlAndDescriptor[0], "/") && !strings.HasPrefix(urlAndDescriptor[0], "//") {
			urlAndDescriptor[0] = baseURL + urlAndDescriptor[0]
			parts[i] = strings.Join(urlAndDescriptor, " ")
		}
	}
	return strings.Join(parts, ", ")
}
