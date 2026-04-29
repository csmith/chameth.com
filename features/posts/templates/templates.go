package templates

import _ "embed"

//go:embed post.html.gotpl
var postTemplateContent string

//go:embed posts.html.gotpl
var postsTemplateContent string
