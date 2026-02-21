package templates

import (
	"html/template"
	"io"
)

var poemTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"article.html.gotpl",
			"poem.html.gotpl",
			"includes/postlink.html.gotpl",
		),
)

type PoemData struct {
	ArticleData
	Poem     []string
	Comments template.HTML
}

func RenderPoem(w io.Writer, poem PoemData) error {
	return poemTemplate.Execute(w, poem)
}
