package templates

import (
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

var listPostsTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(listPostsGotpl))
	return t
}()

var editPostTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(editPostGotpl))
	return t
}()

type ListPostsData struct {
	admintemplates.PageData
	Drafts []PostSummary
	Posts  []PostSummary
}

type PostSummary struct {
	ID    int
	Title string
	Path  string
	Date  string
}

type EditPostData struct {
	admintemplates.PageData
	ID        int
	Title     string
	Path      string
	Date      string
	Content   string
	Format    string
	Published bool
	Media     []PostMediaItem
}

type PostMediaItem struct {
	Path        string
	Title       string
	AltText     string
	Width       *int
	Height      *int
	Role        string
	ContentType string
	MediaID     int
	Variants    []PostMediaVariant
}

type PostMediaVariant struct {
	MediaID     int
	ContentType string
	Width       *int
	Height      *int
}

func RenderListPosts(w http.ResponseWriter, data ListPostsData) error {
	return listPostsTemplate.Execute(w, data)
}

func RenderEditPost(w http.ResponseWriter, data EditPostData) error {
	return editPostTemplate.Execute(w, data)
}
