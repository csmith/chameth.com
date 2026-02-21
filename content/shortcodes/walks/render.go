package walks

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("walks.html.gotpl").ParseFS(templates, "walks.html.gotpl"))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	walks, err := db.GetAllWalks(ctx.Context)
	if err != nil {
		return "", fmt.Errorf("failed to get walks: %w", err)
	}

	var maxDistance float64
	for _, walk := range walks {
		if walk.DistanceKm > maxDistance {
			maxDistance = walk.DistanceKm
		}
	}

	entries := make([]WalkEntry, len(walks))
	for i, walk := range walks {
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
