package list

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"chameth.com/chameth.com/features/shortcodes/common"

	"chameth.com/chameth.com/features/walks"
)

//go:embed *.gotpl
var templates string

var tmpl = template.Must(template.New("walks.html.gotpl").Parse(templates))

func RenderFromText(_ []string, ctx *common.Context) (string, error) {
	allWalks, err := walks.AllWalks(ctx.Context)
	if err != nil {
		return "", fmt.Errorf("failed to get walks: %w", err)
	}

	var maxDistance float64
	for _, w := range allWalks {
		if w.DistanceKm > maxDistance {
			maxDistance = w.DistanceKm
		}
	}

	entries := make([]WalkEntry, len(allWalks))
	for i, walk := range allWalks {
		durationMinutes := int(walk.DurationSeconds / 60)
		barWidth := 0.0
		if maxDistance > 0 {
			barWidth = (walk.DistanceKm / maxDistance) * 100
		}

		var duration strings.Builder
		if hours := durationMinutes / 60; hours > 0 {
			duration.WriteString(strconv.Itoa(hours))
			duration.WriteString(" hour")
			if hours > 1 {
				duration.WriteRune('s')
			}
			duration.WriteRune(' ')
			durationMinutes %= 60
		}
		if durationMinutes != 0 {
			duration.WriteString(strconv.Itoa(durationMinutes))
			duration.WriteString(" minute")
			if durationMinutes > 1 {
				duration.WriteRune('s')
			}
		}

		entries[i] = WalkEntry{
			Date:             walk.StartDate.Format("2006-01-02"),
			DistanceBarWidth: barWidth,
			DistanceKm:       walk.DistanceKm,
			ElevationGainM:   walk.ElevationGainMeters,
			Duration:         strings.TrimSpace(duration.String()),
		}
	}

	return renderTemplate(Data{Walks: entries})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
