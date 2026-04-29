package templates

import _ "embed"

//go:embed list-projects.html.gotpl
var listProjectsGotpl string

//go:embed edit-project.html.gotpl
var editProjectGotpl string
