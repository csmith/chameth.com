package templates

import (
	_ "embed"
)

//go:embed film.html.gotpl
var filmTemplateContent string

//go:embed film_list.html.gotpl
var filmListTemplateContent string
