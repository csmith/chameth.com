package calendar

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"math"
	"time"

	"chameth.com/chameth.com/features/films"
)

//go:embed calendar.html.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("calendar.html.gotpl").ParseFS(templates, "calendar.html.gotpl"))

const weeks = 52

// Shades are fixed CSS classes because the CSP prevents inline styles.
const intensitySteps = 4

var dayNames = [7]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

func buildData(w window, today time.Time, days []films.WatchDay) Data {
	counts := make(map[string]int, len(days))
	var maxCount int
	for _, d := range days {
		counts[d.Date] = d.Count
		maxCount = max(maxCount, d.Count)
	}

	gridStart := startOfWeek(w.start)
	columns := (int(w.end.Sub(gridStart).Hours()/24) + 7) / 7

	rows := make([]DayRow, 7)
	for d := range rows {
		cells := make([]Cell, columns)
		for c := range cells {
			date := gridStart.AddDate(0, 0, c*7+d)
			if date.Before(w.start) || date.After(w.end) || date.After(today) {
				cells[c] = Cell{Inactive: true}
				continue
			}
			cells[c] = buildCell(date, counts[date.Format(time.DateOnly)], maxCount)
		}
		rows[d] = DayRow{DayName: dayNames[d], Cells: cells}
	}

	return Data{Title: w.title, Rows: rows}
}

func buildCell(date time.Time, count, maxCount int) Cell {
	noun := "films"
	if count == 1 {
		noun = "film"
	}
	cell := Cell{Title: fmt.Sprintf("%d %s · %s", count, noun, date.Format("Mon 2 Jan 2006"))}
	if count == 0 {
		cell.Class = "empty"
		return cell
	}

	cell.Class = fmt.Sprintf(
		"filled s%d",
		min(int(math.Round(float64(count)/float64(maxCount)*(intensitySteps-1))), intensitySteps-1),
	)
	return cell
}

func truncateToDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func startOfWeek(t time.Time) time.Time {
	t = truncateToDay(t)
	return t.AddDate(0, 0, -((int(t.Weekday()) - int(time.Monday) + 7) % 7))
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
