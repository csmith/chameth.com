package walkingdistance

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strconv"

	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("walkingdistance.html.gotpl").ParseFS(templates, "walkingdistance.html.gotpl"))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	if len(args) < 3 {
		return "", fmt.Errorf("walkingdistance shortcode requires three arguments: name, distance, svg")
	}

	name := args[0]
	distanceKm, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return "", fmt.Errorf("invalid distance: %w", err)
	}
	svg := args[2]

	totalDistance, err := db.GetTotalDistanceKm(ctx.Context)
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
