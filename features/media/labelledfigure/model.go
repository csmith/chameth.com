package labelledfigure

import "html/template"

type Source struct {
	Src  string
	Type string
}

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
	Sources     []Source
	Src         string
	Description string
	Caption     template.HTML
	Width       int
	Height      int
	Regions     []Region
}
