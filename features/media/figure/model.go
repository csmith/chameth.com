package figure

import (
	"html/template"

	"chameth.com/chameth.com/features/media"
)

type Data struct {
	Class       string
	Sources     []media.Source
	Src         string
	Description string
	Caption     template.HTML
	Width       int
	Height      int
}
