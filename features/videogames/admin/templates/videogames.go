package templates

import (
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
	mediatemplates "chameth.com/chameth.com/features/media/admin/templates"
)

var listVideoGamesTemplate = admintemplates.ParsePage(listVideoGamesGotpl)

var editVideoGameTemplate = admintemplates.ParsePage(editVideoGameGotpl)

var editVideoGameReviewTemplate = admintemplates.ParsePage(editVideoGameReviewGotpl)

type ListVideoGamesData struct {
	admintemplates.PageData
	VideoGames []VideoGameSummary
}

type VideoGameSummary struct {
	ID        int
	Title     string
	Platform  string
	Rating    string
	Published bool
}

type EditVideoGameData struct {
	admintemplates.PageData
	ID        int
	Title     string
	Platform  string
	Overview  string
	Published bool
	Path      string
	Poster    *mediatemplates.MediaItem
	Reviews   []VideoGameReviewSummary
}

type MediaItem = mediatemplates.MediaItem

type VideoGameReviewSummary struct {
	ID               int
	PlayedDate       string
	Rating           string
	Playtime         string
	CompletionStatus string
	Notes            string
	Published        bool
}

type EditVideoGameReviewData struct {
	admintemplates.PageData
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

func RenderListVideoGames(w http.ResponseWriter, data ListVideoGamesData) error {
	return listVideoGamesTemplate.Execute(w, data)
}

func RenderEditVideoGame(w http.ResponseWriter, data EditVideoGameData) error {
	return editVideoGameTemplate.Execute(w, data)
}

func RenderEditVideoGameReview(w http.ResponseWriter, data EditVideoGameReviewData) error {
	return editVideoGameReviewTemplate.Execute(w, data)
}
