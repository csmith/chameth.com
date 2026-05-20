package films

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterContentTypes(rm *routing.Manager) {
	rm.RegisterContentType("film", FilmPage)
	rm.RegisterContentType("film_list", FilmListPage)
}

func RegisterRoutes(rm *routing.Manager) {
	rm.Public.HandleFunc("GET /api/films/search", handleSearch)
}
