package templates

import _ "embed"

//go:embed staticpage.html.gotpl
var staticPageTemplateContent string

//go:embed rawpage.html.gotpl
var rawPageTemplateContent string
