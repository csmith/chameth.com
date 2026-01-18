package postlink

import "html/template"

type Data struct {
	Url     string
	Images  []Image
	Title   string
	Summary template.HTML
}

type Image struct {
	Url         string
	ContentType string
}
