package posts

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterContentTypes(rm *routing.Manager) {
	rm.RegisterContentType("post", PostHandler)
}

func RegisterRoutes(rm *routing.Manager) {
	rm.Public.HandleFunc("GET /posts/{$}", handleList)
}
