package filmreview

import "html/template"

type Data struct {
	Name       string
	PosterPath string
	Rating     int
	Stars      template.HTML
	Date       string
	Rewatch    bool
	Spoiler    bool
	Review     template.HTML
}
