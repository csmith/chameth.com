package sidenote

import "html/template"

type Data struct {
	Title   string
	Content template.HTML
}
