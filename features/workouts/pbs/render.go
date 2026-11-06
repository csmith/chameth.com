package pbs

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"math"

	"chameth.com/chameth.com/features/workouts"
)

//go:embed pbs.html.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("pbs.html.gotpl").ParseFS(templates, "pbs.html.gotpl"))

func buildData(title string, records []record) Data {
	rows := make([]Row, len(records))
	for i, r := range records {
		rows[i] = Row{
			Distance: workouts.FormatDistanceLabel(r.DistanceM),
			Time:     workouts.FormatDuration(r.ElapsedS),
			Pace:     formatPace(r.PaceSPerKm),
			Date:     workouts.FormatDate(r.Date),
		}
	}
	return Data{Title: title, Rows: rows}
}

func formatPace(sPerKm float64) string {
	total := int(math.Round(sPerKm))
	return fmt.Sprintf("%d:%02d/km", total/60, total%60)
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
