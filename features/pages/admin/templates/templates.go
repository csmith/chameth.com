package templates

import _ "embed"

//go:embed list-pages.html.gotpl
var listPagesGotpl string

//go:embed edit-page.html.gotpl
var editPageGotpl string
