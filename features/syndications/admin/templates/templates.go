package templates

import _ "embed"

//go:embed list-syndications.html.gotpl
var listSyndicationsGotpl string

//go:embed edit-syndication.html.gotpl
var editSyndicationGotpl string
