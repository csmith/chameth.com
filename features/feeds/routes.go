package feeds

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Public.HandleFunc("GET /index.xml", handleAllPosts)
	rm.Public.HandleFunc("GET /short.xml", handleShortPosts)
	rm.Public.HandleFunc("GET /long.xml", handleLongPosts)
	rm.Public.HandleFunc("GET /poems/feed.xml", handlePoems)
	rm.Public.HandleFunc("GET /snippets/feed.xml", handleSnippets)
	rm.Public.HandleFunc("GET /films/reviews/feed.xml", handleFilmReviews)
}
