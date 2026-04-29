package templates

import _ "embed"

//go:embed list-goimports.html.gotpl
var listGoImportsGotpl string

//go:embed edit-goimport.html.gotpl
var editGoImportGotpl string
