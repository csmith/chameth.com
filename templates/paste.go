package templates

import (
	"html/template"
	"net/http"
)

var pasteTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"article.html.gotpl",
			"paste.html.gotpl",
		),
)

type PasteData struct {
	ArticleData
	Content  template.HTML
	Language string
	Size     int
}

func RenderPaste(w http.ResponseWriter, paste PasteData) error {
	return pasteTemplate.Execute(w, paste)
}
