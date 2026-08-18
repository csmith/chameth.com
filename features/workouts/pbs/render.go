package pbs

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

//go:embed pbs.html.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("pbs.html.gotpl").ParseFS(templates, "pbs.html.gotpl"))

// activities maps the shortcode's activity argument (and its activity_group
// slug) to the group stored on workouts, plus a display label for the box
// title. Only groups that track segment PBs are listed.
var activities = map[string]struct{ group, label string }{
	"running": {"run", "Running"},
	"run":     {"run", "Running"},
	"cycling": {"cycle", "Cycling"},
	"cycle":   {"cycle", "Cycling"},
	"walking": {"walk", "Walking"},
	"walk":    {"walk", "Walking"},
}

// RenderFromText renders the workoutpbs shortcode: a table of the current
// all-time PB for each segment distance within an activity, e.g.
// {%workoutpbs "running"%}.
func RenderFromText(args []string, ctx *shortcodes.Context) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("workoutpbs requires 1 argument (activity), e.g. \"running\"")
	}
	a, ok := activities[strings.ToLower(args[0])]
	if !ok {
		return "", fmt.Errorf("unknown activity: %s", args[0])
	}

	records, err := workouts.PBsForGroup(ctx.Context, a.group)
	if err != nil {
		return "", fmt.Errorf("failed to get PBs: %w", err)
	}
	if len(records) == 0 {
		return "", nil
	}

	return renderTemplate(buildData(a.label+" PBs", records))
}

func buildData(title string, records []workouts.ActivityRecord) Data {
	rows := make([]Row, len(records))
	for i, r := range records {
		rows[i] = Row{
			Distance: formatDistanceLabel(r.DistanceM),
			Time:     formatDuration(r.ElapsedS),
			Pace:     formatPace(r.PaceSPerKm),
			Date:     r.StartTime.Format("2 Jan 2006"),
		}
	}
	return Data{Title: title, Rows: rows}
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

// formatPace formats a pace in seconds per km as e.g. "4:52/km".
func formatPace(sPerKm float64) string {
	total := int(math.Round(sPerKm))
	return fmt.Sprintf("%d:%02d/km", total/60, total%60)
}

// formatDistanceLabel turns a segment distance in metres into a short,
// human-friendly label, e.g. "5km", "800m", "1 mile".
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
