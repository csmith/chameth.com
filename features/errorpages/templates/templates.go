package templates

import _ "embed"

//go:embed notfound.html.gotpl
var notFoundTemplateContent string

//go:embed servererror.html.gotpl
var serverErrorTemplateContent string
