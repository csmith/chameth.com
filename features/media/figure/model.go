package figure

import "html/template"

type Source struct {
	Src  string
	Type string
}

type Data struct {
	Class       string
	Sources     []Source
	Src         string
	Description string
	Caption     template.HTML
	Width       int
	Height      int
}
