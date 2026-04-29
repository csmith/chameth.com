package templates

import _ "embed"

//go:embed list-pastes.html.gotpl
var listPastesGotpl string

//go:embed edit-paste.html.gotpl
var editPasteGotpl string
