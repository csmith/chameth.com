package filmratingdistribution

import "html/template"

type Data struct {
	Bars       []Bar
	LeftLabel  template.HTML
	RightLabel template.HTML
}

type Bar struct {
	X      int
	Y      int
	Width  int
	Height int
	Title  string
}
