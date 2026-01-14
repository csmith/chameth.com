package templates

import (
	"html/template"
	"net/http"
)

var editVideoGameReviewTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"edit-video-game-review.html.gotpl",
		),
)

type EditVideoGameReviewData struct {
	PageData
	ID               int
	VideoGameID      int
	VideoGameTitle   string
	PlayedDate       string
	Rating           string
	Playtime         string
	CompletionStatus string
	Notes            string
	Published        bool
}

func RenderEditVideoGameReview(w http.ResponseWriter, data EditVideoGameReviewData) error {
	return editVideoGameReviewTemplate.Execute(w, data)
}
