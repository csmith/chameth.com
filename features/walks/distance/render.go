package distance

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"strconv"

	"chameth.com/chameth.com/features/shortcodes"

	"chameth.com/chameth.com/features/walks"
)

//go:embed *.gotpl
var templates string

var tmpl = template.Must(template.New("walkingdistance.html.gotpl").Parse(templates))

func RenderFromText(args []string, ctx *shortcodes.Context) (string, error) {
	if len(args) < 3 {
		return "", fmt.Errorf("walkingdistance shortcode requires three arguments: name, distance, svg")
	}

	name := args[0]
	distanceKm, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return "", fmt.Errorf("invalid distance: %w", err)
	}
	svg := args[2]

	totalDistance, err := walks.TotalDistance(ctx.Context)
	if err != nil {
		return "", fmt.Errorf("failed to get total distance: %w", err)
	}

	timesCompleted := totalDistance / distanceKm
	completedPortion := int(timesCompleted)
	progressPercent := int((timesCompleted - float64(completedPortion)) * 100)

	return renderTemplate(Data{
		Name:            name,
		DistanceKm:      distanceKm,
		SVG:             template.HTML(svg),
		TimesCompleted:  timesCompleted,
		ProgressPercent: progressPercent,
	})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
