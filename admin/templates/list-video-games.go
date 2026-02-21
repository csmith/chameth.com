package templates

import (
	"html/template"
	"net/http"
)

var listVideoGamesTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"list-video-games.html.gotpl",
		),
)

type ListVideoGamesData struct {
	PageData
	VideoGames []VideoGameSummary
}

type VideoGameSummary struct {
	ID        int
	Title     string
	Platform  string
	Rating    string
	Published bool
}

func RenderListVideoGames(w http.ResponseWriter, data ListVideoGamesData) error {
	return listVideoGamesTemplate.Execute(w, data)
}
