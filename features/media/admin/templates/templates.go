package templates

import _ "embed"

//go:embed media.html.gotpl
var mediaGotpl string

//go:embed edit-media.html.gotpl
var editMediaGotpl string

//go:embed edit-media-relations.html.gotpl
var editMediaRelationsGotpl string
