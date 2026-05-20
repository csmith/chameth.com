package snippets

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterContentTypes(rm *routing.Manager) {
	rm.RegisterContentType("snippet", SnippetHandler)
}

func RegisterRoutes(rm *routing.Manager) {
	rm.Public.HandleFunc("GET /snippets/{$}", handleList)
}
