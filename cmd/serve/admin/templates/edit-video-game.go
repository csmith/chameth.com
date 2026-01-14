package templates

import (
	"html/template"
	"net/http"
)

var editVideoGameTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"edit-video-game.html.gotpl",
		),
)

type EditVideoGameData struct {
	PageData
	ID        int
	Title     string
	Platform  string
	Overview  string
	Published bool
	Path      string
	Poster    *MediaItem
	Reviews   []VideoGameReviewSummary
}

type VideoGameReviewSummary struct {
	ID               int
	PlayedDate       string
	Rating           string
	Playtime         string
	CompletionStatus string
	Notes            string
	Published        bool
}

func RenderEditVideoGame(w http.ResponseWriter, data EditVideoGameData) error {
	return editVideoGameTemplate.Execute(w, data)
}
