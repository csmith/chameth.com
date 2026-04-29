package templates

import _ "embed"

//go:embed list-poems.html.gotpl
var listPoemsGotpl string

//go:embed edit-poem.html.gotpl
var editPoemGotpl string
