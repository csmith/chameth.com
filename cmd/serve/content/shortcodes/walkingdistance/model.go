package walkingdistance

import "html/template"

type Data struct {
	Name            string
	DistanceKm      float64
	SVG             template.HTML
	TimesCompleted  float64
	ProgressPercent int
}
