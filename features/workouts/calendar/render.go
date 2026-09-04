package calendar

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"
)

//go:embed calendar.html.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("calendar.html.gotpl").ParseFS(templates, "calendar.html.gotpl"))

const weeks = 16

// Shades are fixed CSS classes because the CSP prevents inline styles.
const intensitySteps = 4

var scalingTypes = map[string]bool{
	"run":   true,
	"cycle": true,
	"walk":  true,
}

var dayNames = [7]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

type dayBucket struct {
	groups map[string]float64
	count  int
}

func indexDays(entries []dayEntry) map[string]*dayBucket {
	byDay := make(map[string]*dayBucket, len(entries))
	for _, e := range entries {
		byDay[e.Date] = &dayBucket{groups: e.Distances, count: e.Count}
	}
	return byDay
}

// Keep zero-distance types in the map so they still appear in the legend.
func perTypeMaxDay(byDay map[string]*dayBucket) map[string]float64 {
	max := map[string]float64{}
	for _, b := range byDay {
		for t, dist := range b.groups {
			if prev, ok := max[t]; !ok || dist > prev {
				max[t] = dist
			}
		}
	}
	return max
}

func buildData(title string, start, end, rangeStart, rangeEnd time.Time, entries []dayEntry) Data {
	byDay := indexDays(entries)
	maxByType := perTypeMaxDay(byDay)

	gridStart := startOfWeek(start)
	days := int(end.Sub(gridStart).Hours()/24) + 1
	columns := (days + 6) / 7

	rows := make([]DayRow, 7)
	for d := range rows {
		cells := make([]Cell, columns)
		for w := range cells {
			date := gridStart.AddDate(0, 0, w*7+d)
			cells[w] = buildCell(date, start, end, rangeStart, rangeEnd, byDay, maxByType)
		}
		rows[d] = DayRow{DayName: dayNames[d], Cells: cells}
	}

	return Data{
		Title:  title,
		Rows:   rows,
		Legend: buildLegend(maxByType),
	}
}

func buildCell(date, start, end, rangeStart, rangeEnd time.Time, byDay map[string]*dayBucket, maxByType map[string]float64) Cell {
	if date.Before(start) || date.After(end) {
		return Cell{Inactive: true}
	}
	cell := Cell{Title: dayTitle(0, nil, date)}
	if !rangeStart.IsZero() && !date.Before(rangeStart) && !date.After(rangeEnd) {
		cell.InRange = true
	}
	bucket, ok := byDay[date.Format("2006-01-02")]
	if !ok {
		return cell
	}
	cell.Title = dayTitle(bucket.count, bucket.groups, date)
	cell.Stripes = buildStripes(bucket.groups, maxByType)
	return cell
}

func buildStripes(groups map[string]float64, maxByType map[string]float64) []Stripe {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	stripes := make([]Stripe, 0, len(keys))
	for _, k := range keys {
		var ratio float64
		if max := maxByType[k]; max > 0 {
			ratio = groups[k] / max
		}
		stripes = append(stripes, Stripe{Class: stripeClass(k, ratio)})
	}
	return stripes
}

func stripeClass(group string, ratio float64) string {
	if !scalingTypes[group] {
		return group
	}
	return fmt.Sprintf("%s s%d", group, intensityStep(ratio))
}

func intensityStep(ratio float64) int {
	step := int(math.Round(ratio * (intensitySteps - 1)))
	return min(step, intensitySteps-1)
}

func buildLegend(maxByType map[string]float64) []LegendRow {
	var hasOther bool
	named := make([]string, 0, len(maxByType))
	for k := range maxByType {
		if k == "other" {
			hasOther = true
			continue
		}
		named = append(named, k)
	}
	sort.Strings(named)
	if hasOther {
		named = append(named, "other")
	}

	rows := make([]LegendRow, 0, len(named))
	for _, k := range named {
		var swatches []string
		if scalingTypes[k] {
			swatches = make([]string, intensitySteps)
			for i := range swatches {
				swatches[i] = fmt.Sprintf("%s s%d", k, i)
			}
		} else {
			swatches = []string{k}
		}
		rows = append(rows, LegendRow{Label: formatSport(k), Swatches: swatches})
	}
	return rows
}

func dayTitle(n int, groups map[string]float64, date time.Time) string {
	noun := "activities"
	if n == 1 {
		noun = "activity"
	}
	s := fmt.Sprintf("%d %s \u00b7 %s", n, noun, date.Format("Mon 2 Jan 2006"))
	if len(groups) == 0 {
		return s
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %s", formatSport(k), formatDistanceKm(groups[k])))
	}
	return s + " \u00b7 " + strings.Join(parts, ", ")
}

// Distances are not comparable across the activities folded into "other".
func calendarGroup(group string) string {
	if scalingTypes[group] {
		return group
	}
	return "other"
}

func truncateToDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func startOfWeek(t time.Time) time.Time {
	t = truncateToDay(t)
	daysFromMonday := (int(t.Weekday()) - int(time.Monday) + 7) % 7
	return t.AddDate(0, 0, -daysFromMonday)
}

func formatSport(group string) string {
	if group == "" {
		return "Other"
	}
	r := []rune(group)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func formatDistanceKm(m float64) string {
	switch {
	case m <= 0:
		return "—"
	case m < 1000:
		return fmt.Sprintf("%d m", int(math.Round(m)))
	default:
		return fmt.Sprintf("%.2f km", m/1000)
	}
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
