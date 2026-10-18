package films

import (
	"chameth.com/chameth.com/features/routing"
	"chameth.com/chameth.com/features/shortcodes"
)

func RegisterContentTypes(rm *routing.Manager, mgr *shortcodes.Manager) {
	rm.RegisterContentType("film", filmPageHandler(mgr))
	rm.RegisterContentType("film_list", filmListPage)
}

func RegisterRoutes(rm *routing.Manager) {
	rm.Public.HandleFunc("GET /api/films/search", handleSearch)
}
