package calendar

// Data is the view model for the workoutcalendar shortcode.
type Data struct {
	Title  string
	Rows   []DayRow
	Legend []LegendRow
}

// DayRow is one weekday row of the calendar, spanning every week column.
// DayName is the abbreviated weekday ("Mon", "Tue", ...) shown as the row
// header. Cells are ordered oldest-week first so the most recent week
// lands in the rightmost column.
type DayRow struct {
	DayName string
	Cells   []Cell
}

// Cell is one day. Stripes is empty for a day with no workouts; otherwise
// it carries one stripe per activity type present that day (multiple
// activities of the same type collapse into a single stripe whose shade
// reflects the type's summed distance). Title is the tooltip text — the
// workout count, the long-form date and a per-type distance breakdown.
// Inactive marks cells outside the calendar's date span (the Monday-padded
// edges of the grid); InRange marks cells inside a requested date range,
// which are rendered with a highlight.
type Cell struct {
	Title    string
	Stripes  []Stripe
	Inactive bool
	InRange  bool
}

// Stripe is one horizontal band within a filled cell: the CSS classes
// picking the activity type's colour and its quantised intensity step
// (defined in style.public.css). All stripes within a cell take an equal
// share of the height (set via flex:1 in CSS) — the shade carries the
// intensity encoding, so weighting by distance would double-encode.
type Stripe struct {
	Class string
}

// LegendRow is one entry in the legend below the calendar: an activity
// type label and swatch classes spanning that type's intensity scale so
// the cell shades can be decoded. Only types that actually appear in the
// visible window are listed.
type LegendRow struct {
	Label    string
	Swatches []string
}
