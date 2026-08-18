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

	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/workouts"
)

//go:embed calendar.html.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("calendar.html.gotpl").ParseFS(templates, "calendar.html.gotpl"))

// weeks is how many weeks the calendar spans: the exact window when no
// date range is given, and the minimum window (ending at the range's end)
// when one is.
const weeks = 16

// intensitySteps is how many discrete shades each activity type's
// intensity scale is quantised into, s0 (dimmest) through s3 (the type's
// full tint) — the same steps the legend shows, so every cell shade
// matches a legend swatch exactly. The colours themselves live in
// style.public.css as fixed type × step classes; the site's CSP blocks
// inline style attributes, so they can't be computed per render.
const intensitySteps = 4

// scalingTypes are the activity types whose stripes are shaded by
// distance: the day's type total divided by the type's max single-day
// total across the visible window, quantised into intensitySteps shades.
// Every other activity group (gym, rowing, dance, ...) folds into the
// catch-all "other" bucket, which renders at a single full tint —
// distance is meaningless across the activity types it folds together.
// The palette (defined in style.public.css) gives run/cycle/walk/other
// distinct hues; walk's ramp starts lighter than the others because its
// muted grey-brown would otherwise fade into the empty-cell background.
var scalingTypes = map[string]bool{
	"run":   true,
	"cycle": true,
	"walk":  true,
}

// dayNames is the row labels for the calendar, Monday-first to match the
// column anchoring. The abbreviated form keeps the row header narrow so
// the cell grid dominates the visual.
var dayNames = [7]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

// RenderFromText renders the workoutcalendar shortcode. With no arguments
// it shows the last 16 weeks of activity. With a start and end date
// (YYYY-MM-DD) it shows a window ending at the end date and extended
// back to the earlier of the start date or 16 weeks, with the requested
// date range highlighted within it.
func RenderFromText(args []string, ctx *shortcodes.Context) (string, error) {
	var title string
	var start, end, rangeStart, rangeEnd time.Time

	switch len(args) {
	case 0:
		now := time.Now().UTC()
		end = truncateToDay(now)
		start = startOfWeek(now).AddDate(0, 0, -(weeks-1)*7)
		title = "Activity calendar"
	case 2:
		var err error
		rangeStart, err = time.Parse("2006-01-02", args[0])
		if err != nil {
			return "", fmt.Errorf("invalid start date: %s (expected YYYY-MM-DD)", args[0])
		}
		rangeEnd, err = time.Parse("2006-01-02", args[1])
		if err != nil {
			return "", fmt.Errorf("invalid end date: %s (expected YYYY-MM-DD)", args[1])
		}
		end = rangeEnd
		start = rangeEnd.AddDate(0, 0, -7*weeks)
		if rangeStart.Before(start) {
			start = rangeStart
		}
		title = fmt.Sprintf("Activity calendar %s - %s",
			rangeStart.Format("2006-01-02"), rangeEnd.Format("2006-01-02"))
	default:
		return "", fmt.Errorf("workoutcalendar requires 0 or 2 arguments (start_date, end_date) in YYYY-MM-DD format")
	}

	entries, err := workouts.WorkoutDayEntries(ctx.Context, start, end.AddDate(0, 0, 1))
	if err != nil {
		return "", fmt.Errorf("failed to get workout days: %w", err)
	}

	return renderTemplate(buildData(title, start, end, rangeStart, rangeEnd, entries))
}

// dayBucket is one day's rolled-up workout data: per-activity-group
// distance totals (driving both the cell's stripes and the per-type
// intensity scale) and the workout count (driving the tooltip).
type dayBucket struct {
	groups map[string]float64
	count  int
}

// indexWorkoutDays buckets entries by YYYY-MM-DD, summing distances per
// activity group within each day. Run-grouped workouts count only their
// running distance (run_distance_m), so the walking warmup/cooldown
// intervals of a run don't inflate the run shade; running intervals
// inside other groups (e.g. a couch-to-5k walk) stay attributed to that
// group's bucket. Unknown activity groups are normalised to "other" via
// calendarGroup so they aggregate into a single catch-all stripe + legend
// entry rather than one per ad-hoc name.
func indexWorkoutDays(entries []workouts.WorkoutDayEntry) map[string]*dayBucket {
	byDay := map[string]*dayBucket{}
	for _, e := range entries {
		key := e.StartTime.Format("2006-01-02")
		bucket, ok := byDay[key]
		if !ok {
			bucket = &dayBucket{groups: map[string]float64{}}
			byDay[key] = bucket
		}
		dist := e.DistanceM
		if e.ActivityGroup == "run" && e.RunDistanceM != nil {
			dist = *e.RunDistanceM
		}
		bucket.groups[calendarGroup(e.ActivityGroup)] += dist
		bucket.count++
	}
	return byDay
}

// perTypeMaxDay returns, for each activity type, the largest single-day
// distance total across the indexed days. The result is the denominator
// of the per-type intensity scale: each stripe's shade is its day-total
// divided by this max, so the longest day for each type hits the fullest
// shade regardless of how the surrounding window is scoped.
//
// A type appears in the result iff any bucket has the type key, even when
// the distance is zero (e.g. gym) — without this, the legend would drop
// zero-distance types entirely even though their cells render at full
// tint. Tracking presence explicitly (rather than relying on
// `dist > max[t]`, which fails for dist=0 since the zero value of an
// absent key also compares equal) keeps the legend in sync with the grid.
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

// buildData assembles the calendar as 7 weekday rows × N week columns,
// oldest week first so the most recent week lands in the rightmost
// column. The grid is anchored on the Monday on or before the window
// start so every column is a clean Mon–Sun span; days padded outside
// [start, end] by that anchoring render as inactive cells. When a date
// range was requested, cells within it are flagged for highlighting.
func buildData(title string, start, end, rangeStart, rangeEnd time.Time, entries []workouts.WorkoutDayEntry) Data {
	byDay := indexWorkoutDays(entries)
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

// buildCell assembles one calendar cell for the given date: an inactive
// cell outside the window, an empty cell with the zero-count tooltip if
// no workouts landed on it, or a filled cell with stripes otherwise.
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

// buildStripes assembles one stripe per activity type present that day,
// alphabetically so multi-type days render consistently. Each stripe
// carries the type's class plus the quantised intensity step for the
// day's type total against the type's max single-day total across the
// visible window.
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

// stripeClass returns the CSS classes for one stripe: the activity type,
// plus an intensity step class for types that are shaded by distance.
func stripeClass(group string, ratio float64) string {
	if !scalingTypes[group] {
		return group
	}
	return fmt.Sprintf("%s s%d", group, intensityStep(ratio))
}

// intensityStep buckets a 0–1 intensity ratio into one of intensitySteps
// discrete steps (0 = dimmest, intensitySteps-1 = full tint).
func intensityStep(ratio float64) int {
	step := int(math.Round(ratio * (intensitySteps - 1)))
	return min(step, intensitySteps-1)
}

// buildLegend produces one row per activity type present in the visible
// window, named sports alphabetically with "other" last. Scaling types
// list one swatch class per intensity step; the catch-all lists its
// single full-tint swatch.
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

// dayTitle formats a cell's tooltip: the workout count, the long-form
// date, and a per-type distance breakdown for days with activities.
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

// calendarGroup normalises an activity_group for the calendar's per-type
// bucketing. Run/cycle/walk pass through unchanged; everything else (gym,
// rowing, dance, ...) folds into "other" so the catch-all bucket
// aggregates every unrecognised group into a single stripe + legend
// entry. Without this, a gym session and a dance class on the same day
// would render as two separate dark stripes rather than one.
func calendarGroup(group string) string {
	if scalingTypes[group] {
		return group
	}
	return "other"
}

// truncateToDay returns midnight UTC on the given day.
func truncateToDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// startOfWeek returns the midnight UTC Monday on or before the given
// day, so every grid column is a clean Mon–Sun span.
func startOfWeek(t time.Time) time.Time {
	t = truncateToDay(t)
	daysFromMonday := (int(t.Weekday()) - int(time.Monday) + 7) % 7
	return t.AddDate(0, 0, -daysFromMonday)
}

// formatSport turns an activity type slug into its legend/tooltip label.
func formatSport(group string) string {
	if group == "" {
		return "Other"
	}
	r := []rune(group)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// formatDistanceKm formats a distance for a cell tooltip: metres below a
// kilometre, two-decimal kilometres above, and an em dash for workouts
// with no distance at all (e.g. gym).
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
