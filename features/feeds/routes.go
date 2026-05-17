package feeds

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /index.xml", handleAllPosts)
	mux.HandleFunc("GET /short.xml", handleShortPosts)
	mux.HandleFunc("GET /long.xml", handleLongPosts)
	mux.HandleFunc("GET /poems/feed.xml", handlePoems)
	mux.HandleFunc("GET /snippets/feed.xml", handleSnippets)
	mux.HandleFunc("GET /films/reviews/feed.xml", handleFilmReviews)
}
