package longest

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"math"
	"strconv"
	"strings"

	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/workouts"
)

//go:embed longest.html.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("longest.html.gotpl").ParseFS(templates, "longest.html.gotpl"))

// milestone is a notable round distance that the longest run or ride is
// measured against.
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

// RenderRunFromText renders the longestrun shortcode: the longest run ever
// recorded, with progress towards the next running milestone.
func RenderRunFromText(_ []string, ctx *shortcodes.Context) (string, error) {
	w, err := workouts.FurthestRun(ctx.Context)
	if err != nil {
		return "", fmt.Errorf("failed to get furthest run: %w", err)
	}
	if w == nil {
		return "", nil
	}
	return renderTemplate(buildData("Longest run", w, runMilestones))
}

// RenderCycleFromText renders the longestcycle shortcode: the longest bike
// ride ever recorded, with progress towards the next cycling milestone.
func RenderCycleFromText(_ []string, ctx *shortcodes.Context) (string, error) {
	w, err := workouts.FurthestCycle(ctx.Context)
	if err != nil {
		return "", fmt.Errorf("failed to get furthest cycle: %w", err)
	}
	if w == nil {
		return "", nil
	}
	return renderTemplate(buildData("Longest bike ride", w, cycleMilestones))
}

func buildData(title string, w *workouts.FurthestWorkout, milestones []milestone) Data {
	next := nextMilestone(milestones, w.DistanceM)
	return Data{
		Title:           title,
		Distance:        formatKm(w.DistanceM),
		Date:            w.StartTime.Format("2 Jan 2006"),
		NextMilestone:   next.Label,
		ProgressPercent: int(math.Min(w.DistanceM/next.DistanceM*100, 100)),
	}
}

// nextMilestone returns the first milestone beyond the given distance, or
// the final milestone if the distance has already reached them all.
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
