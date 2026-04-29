package templates

import (
	"fmt"
	"html/template"
	"io"

	parenttemplates "chameth.com/chameth.com/templates"
)

var filmTemplate = func() *template.Template {
	pageContent, err := parenttemplates.FS.ReadFile("page.html.gotpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read page.html.gotpl: %v", err))
	}
	t := template.Must(template.New("page.html.gotpl").Parse(string(pageContent)))
	template.Must(t.Parse(filmTemplateContent))
	return t
}()

var filmListTemplate = func() *template.Template {
	pageContent, err := parenttemplates.FS.ReadFile("page.html.gotpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read page.html.gotpl: %v", err))
	}
	t := template.Must(template.New("page.html.gotpl").Parse(string(pageContent)))
	template.Must(t.Parse(filmListTemplateContent))
	return t
}()

type FilmData struct {
	parenttemplates.PageData
	Title         string
	Year          string
	TMDBID        *int
	Overview      template.HTML
	Reviews       []FilmReviewData
	TimesWatched  int
	AverageRating int
	PosterPath    string
	FilmLists     []int
}

type FilmReviewData struct {
	WatchedDate string
	Rating      int
	IsRewatch   bool
	HasSpoilers bool
	Content     template.HTML
}

type FilmListData struct {
	parenttemplates.PageData
	ListTitle   string
	Description template.HTML
	Entries     []FilmListItem
}

type FilmListItem struct {
	Position     int
	PosterPath   string
	FilmPath     string
	Title        string
	Year         string
	TimesWatched int
	RatingText   string
	Rating       int
	LastWatched  string
}

func RenderFilm(w io.Writer, film FilmData) error {
	return filmTemplate.Execute(w, film)
}

func RenderFilmList(w io.Writer, filmList FilmListData) error {
	return filmListTemplate.Execute(w, filmList)
}
