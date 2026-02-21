package templates

import (
	"html/template"
	"io"
)

var printsTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"prints.html.gotpl",
		),
)

type PrintsData struct {
	PageData
	Prints []PrintDetails
}

type PrintDetails struct {
	Name        string
	Description string
	RenderPath  string
	PreviewPath string
	Links       []PrintLink
}

type PrintLink struct {
	Name    string
	Address string
}

func RenderPrints(w io.Writer, data PrintsData) error {
	return printsTemplate.Execute(w, data)
}
