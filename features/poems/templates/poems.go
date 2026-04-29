package templates

import (
	"fmt"
	"html/template"
	"io"

	parenttemplates "chameth.com/chameth.com/templates"
)

var poemTemplate = func() *template.Template {
	pageContent, err := parenttemplates.FS.ReadFile("page.html.gotpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read page.html.gotpl: %v", err))
	}
	t := template.Must(template.New("page.html.gotpl").Parse(string(pageContent)))

	articleContent, err := parenttemplates.FS.ReadFile("article.html.gotpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read article.html.gotpl: %v", err))
	}
	template.Must(t.Parse(string(articleContent)))

	template.Must(t.Parse(poemTemplateContent))
	return t
}()

type PoemData struct {
	parenttemplates.ArticleData
	Poem     []string
	Comments template.HTML
}

func RenderPoem(w io.Writer, data PoemData) error {
	return poemTemplate.Execute(w, data)
}
