package walkingspeed

import "html/template"

type Data struct {
	SVG template.HTML
}

type point struct {
	X     int
	Y     int
	Title string
}
