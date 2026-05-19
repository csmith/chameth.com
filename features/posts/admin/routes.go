package admin

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Admin.HandleFunc("GET /posts", ListPostsHandler())
	rm.Admin.HandleFunc("POST /posts", CreatePostHandler())
	rm.Admin.HandleFunc("GET /posts/edit/{id}", EditPostHandler())
	rm.Admin.HandleFunc("POST /posts/edit/{id}", UpdatePostHandler())
	rm.Admin.HandleFunc("POST /posts/generate-wordcloud/{id}", GenerateWordcloudHandler())
}
