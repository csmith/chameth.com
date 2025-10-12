package templates

import (
	"html/template"
	"io"
)

var poemTemplate = template.Must(
	template.
		New("page.html.gotpl").
		Funcs(funcMap).
		ParseFS(
			templates,
			"page.html.gotpl",
			"article.html.gotpl",
			"poem.html.gotpl",
		),
)

type PoemData struct {
	ArticleData
	Poem     string
	Comments string
}

func RenderPoem(w io.Writer, poem PoemData) error {
	return poemTemplate.Execute(w, poem)
}
