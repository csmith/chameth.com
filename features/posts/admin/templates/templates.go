package templates

import _ "embed"

//go:embed list-posts.html.gotpl
var listPostsGotpl string

//go:embed edit-post.html.gotpl
var editPostGotpl string
