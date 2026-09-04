package longest

import (
	"bytes"
	"embed"
	"html/template"
	"math"
	"strconv"
	"strings"
)

//go:embed longest.html.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("longest.html.gotpl").ParseFS(templates, "longest.html.gotpl"))

type milestone struct {
	Label     string
	DistanceM float64
}

var runMilestones = []milestone{
	{Label: "5k", DistanceM: 5000},
	{Label: "10k", DistanceM: 10000},
	{Label: "half marathon", DistanceM: 21097.5},
	{Label: "marathon", DistanceM: 42195},
}

var cycleMilestones = []milestone{
	{Label: "50k", DistanceM: 50000},
	{Label: "century", DistanceM: 100000},
	{Label: "imperial century", DistanceM: 160934},
	{Label: "double century", DistanceM: 200000},
}

func buildData(title string, r *record, milestones []milestone) Data {
	next := nextMilestone(milestones, r.DistanceM)
	return Data{
		Title:           title,
		Distance:        formatKm(r.DistanceM),
		Date:            r.StartTime.Format("2 Jan 2006"),
		NextMilestone:   next.Label,
		ProgressPercent: int(math.Min(r.DistanceM/next.DistanceM*100, 100)),
	}
}

func nextMilestone(milestones []milestone, distanceM float64) milestone {
	for _, m := range milestones {
		if distanceM < m.DistanceM {
			return m
		}
	}
	return milestones[len(milestones)-1]
}

func formatKm(m float64) string {
	km := m / 1000
	s := strconv.FormatFloat(km, 'f', 1, 64)
	s = strings.TrimSuffix(s, ".0")
	return s + "km"
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
