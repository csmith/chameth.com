package labelledfigure

import (
	"html/template"

	"chameth.com/chameth.com/features/media"
)

type Region struct {
	X        int
	Y        int
	W        int
	H        int
	CenterX  int
	CenterY  int
	Colour   string
	Label    string
	FontSize int
}

type Data struct {
	Sources     []media.Source
	Src         string
	Description string
	Caption     template.HTML
	Width       int
	Height      int
	Regions     []Region
}
