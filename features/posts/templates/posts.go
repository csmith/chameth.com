package templates

import (
	"fmt"
	"html/template"
	"io"

	parenttemplates "chameth.com/chameth.com/templates"
)

var postTemplate = func() *template.Template {
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

	template.Must(t.Parse(postTemplateContent))
	return t
}()

var postsTemplate = func() *template.Template {
	pageContent, err := parenttemplates.FS.ReadFile("page.html.gotpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read page.html.gotpl: %v", err))
	}
	t := template.Must(template.New("page.html.gotpl").Parse(string(pageContent)))

	template.Must(t.Parse(postsTemplateContent))
	return t
}()

type PostData struct {
	parenttemplates.ArticleData
	PostContent template.HTML
	PostFormat  string
}

type PostsData struct {
	parenttemplates.PageData
	Posts []string
}

func RenderPost(w io.Writer, data PostData) error {
	return postTemplate.Execute(w, data)
}

func RenderPosts(w io.Writer, data PostsData) error {
	return postsTemplate.Execute(w, data)
}
