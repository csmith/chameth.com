package postlink

import "html/template"

type Data struct {
	Url     string
	Images  []Image
	Title   string
	Summary template.HTML
}

func (d Data) Alt() string {
	for _, img := range d.Images {
		if img.Alt != "" {
			return img.Alt
		}
	}
	return "Lead image for " + d.Title
}

type Image struct {
	Url         string
	ContentType string
	Alt         string
}
