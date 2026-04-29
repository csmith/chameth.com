package templates

import _ "embed"

//go:embed snippet.html.gotpl
var snippetTemplateContent string

//go:embed snippets.html.gotpl
var snippetsTemplateContent string
