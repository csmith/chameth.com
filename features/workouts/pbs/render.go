package pbs

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"math"
	"strconv"
	"strings"
	"time"
)

//go:embed pbs.html.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("pbs.html.gotpl").ParseFS(templates, "pbs.html.gotpl"))

func buildData(title string, records []record) Data {
	rows := make([]Row, len(records))
	for i, r := range records {
		rows[i] = Row{
			Distance: formatDistanceLabel(r.DistanceM),
			Time:     formatDuration(r.ElapsedS),
			Pace:     formatPace(r.PaceSPerKm),
			Date:     formatDate(r.Date),
		}
	}
	return Data{Title: title, Rows: rows}
}

func formatDate(date string) string {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return d.Format("2 Jan 2006")
}

func formatDuration(seconds float64) string {
	total := int(math.Round(seconds))
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func formatPace(sPerKm float64) string {
	total := int(math.Round(sPerKm))
	return fmt.Sprintf("%d:%02d/km", total/60, total%60)
}

func formatDistanceLabel(m float64) string {
	if m == 1609 {
		return "1 mile"
	}
	if m < 1000 {
		return fmt.Sprintf("%dm", int(math.Round(m)))
	}
	km := m / 1000
	if km == math.Trunc(km) {
		return fmt.Sprintf("%dkm", int(km))
	}
	s := strconv.FormatFloat(km, 'f', 1, 64)
	s = strings.TrimSuffix(s, ".0")
	return s + "km"
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
