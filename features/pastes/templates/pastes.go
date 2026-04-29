package templates

import (
	"fmt"
	"html/template"
	"io"

	parenttemplates "chameth.com/chameth.com/templates"
)

var pasteTemplate = func() *template.Template {
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

	template.Must(t.Parse(pasteTemplateContent))
	return t
}()

type PasteData struct {
	parenttemplates.ArticleData
	Content  template.HTML
	Language string
	Size     int
}

func RenderPaste(w io.Writer, data PasteData) error {
	return pasteTemplate.Execute(w, data)
}
