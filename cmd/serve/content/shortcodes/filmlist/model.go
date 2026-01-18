package filmlist

import "html/template"

type Film struct {
	ID         int
	Title      string
	Year       string
	PosterPath string
	Path       string
}

type Data struct {
	ID          int
	Title       string
	Description template.HTML
	Path        string
	Count       int
	Films       []Film
}
