package templates

import _ "embed"

//go:embed list-snippets.html.gotpl
var listSnippetsGotpl string

//go:embed edit-snippet.html.gotpl
var editSnippetGotpl string
