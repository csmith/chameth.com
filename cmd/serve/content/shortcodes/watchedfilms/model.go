package watchedfilms

import "html/template"

type Film struct {
	Title      string
	PosterPath string
	Path       string
	Date       string
	Stars      template.HTML
}

type Data struct {
	Films []Film
}
