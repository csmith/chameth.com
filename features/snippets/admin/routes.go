package admin

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Admin.HandleFunc("GET /snippets", ListSnippetsHandler())
	rm.Admin.HandleFunc("POST /snippets", CreateSnippetHandler())
	rm.Admin.HandleFunc("GET /snippets/edit/{id}", EditSnippetHandler())
	rm.Admin.HandleFunc("POST /snippets/edit/{id}", UpdateSnippetHandler())
}
