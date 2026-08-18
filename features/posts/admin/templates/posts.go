package templates

import (
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
	"chameth.com/chameth.com/features/media"
)

var listPostsTemplate = admintemplates.ParsePage(listPostsGotpl)
var editPostTemplate = admintemplates.ParsePage(editPostGotpl)

type ListPostsData = admintemplates.ListData[PostSummary]

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
	Media     []media.GroupedMedia
}

func RenderListPosts(w http.ResponseWriter, data ListPostsData) error {
	return listPostsTemplate.Execute(w, data)
}

func RenderEditPost(w http.ResponseWriter, data EditPostData) error {
	return editPostTemplate.Execute(w, data)
}
