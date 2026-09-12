package parkruns

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strconv"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed *.gotpl
var templates embed.FS

var statsTmpl = template.Must(template.New("parkrunstats.html.gotpl").ParseFS(templates, "parkrunstats.html.gotpl"))
var tableTmpl = template.Must(template.New("parkruns.html.gotpl").ParseFS(templates, "parkruns.html.gotpl"))

func renderStats(_ []string, d statsData, _ *shortcodes.Context) (string, error) {
	best := Box{Title: "Best time", Value: "—"}
	if d.Best != nil {
		best.Value = formatResultTime(d.Best.Time)
		best.Detail = formatBestDetail(d.Best)
	}

	data := StatsData{
		Total:  Box{Title: "Total parkruns", Value: strconv.Itoa(d.Total)},
		Venues: Box{Title: "Total venues", Value: strconv.Itoa(d.Venues)},
		Best:   best,
	}
	return renderTemplate(statsTmpl, data)
}

func renderTable(args []string, runs []runRecord, _ *shortcodes.Context) (string, error) {
	start, end, err := parseRange(args)
	if err != nil {
		return "", err
	}
	if len(runs) == 0 {
		return "", nil
	}

	title := "Parkruns"
	if !start.IsZero() {
		title = fmt.Sprintf("Parkruns %s – %s", start.Format("2 Jan"), end.Format("2 Jan 2006"))
	}

	rows := make([]TableRow, len(runs))
	for i, r := range runs {
		rows[i] = TableRow{
			Date:   formatDate(r.Date),
			Event:  r.Name,
			Time:   formatResultTime(r.GunTime, r.ChipTime),
			Pos:    formatOrdinal(r.PosOverall),
			AgePos: formatOrdinal(r.PosAge),
		}
	}
	return renderTemplate(tableTmpl, TableData{Title: title, Rows: rows})
}

func formatBestDetail(best *bestRun) string {
	detail := best.Location
	if date := formatDate(best.Date); date != "" {
		if detail != "" {
			detail += ", "
		}
		detail += date
	}
	return detail
}

func formatDate(date string) string {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return d.Format("2 Jan 2006")
}

// formatResultTime renders the first recorded time, trimming a leading
// zero hour; recorded times are HH:MM:SS.
func formatResultTime(times ...string) string {
	for _, t := range times {
		if t == "" {
			continue
		}
		if parsed, err := time.Parse("15:04:05", t); err == nil {
			if h := parsed.Hour(); h > 0 {
				return fmt.Sprintf("%d:%02d:%02d", h, parsed.Minute(), parsed.Second())
			}
			return fmt.Sprintf("%d:%02d", parsed.Minute(), parsed.Second())
		}
		return t
	}
	return "—"
}

func formatOrdinal(n *int) string {
	if n == nil {
		return "—"
	}

	v := *n
	suffix := "th"
	if v%100 >= 11 && v%100 <= 13 {
		return strconv.Itoa(v) + "th"
	}
	switch v % 10 {
	case 1:
		suffix = "st"
	case 2:
		suffix = "nd"
	case 3:
		suffix = "rd"
	}
	return strconv.Itoa(v) + suffix
}

func renderTemplate(t *template.Template, data any) (string, error) {
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
