package feeds

import (
	"chameth.com/chameth.com/features/routing"
	"chameth.com/chameth.com/features/shortcodes"
)

func RegisterRoutes(rm *routing.Manager, mgr *shortcodes.Manager) {
	builderCookieCipher()
	rm.Public.HandleFunc("GET /index.xml", handleAllPosts)
	rm.Public.HandleFunc("GET /short.xml", handleShortPosts)
	rm.Public.HandleFunc("GET /long.xml", handleLongPosts)
	rm.Public.HandleFunc("GET /poems/feed.xml", handlePoems)
	rm.Public.HandleFunc("GET /snippets/feed.xml", handleSnippets)
	rm.Public.HandleFunc("GET /films/reviews/feed.xml", handleFilmReviews(mgr))
	rm.Public.HandleFunc("GET /feeds/posts/build/", handleRelatedPostsBuilder)
	rm.Public.HandleFunc("POST /feeds/posts/build/", handleRelatedPostsBuilderAction)
	rm.Public.HandleFunc("GET /feeds/posts/", handleRelatedPostsFeed)
}
