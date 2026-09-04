package calendar

type Data struct {
	Title  string
	Rows   []DayRow
	Legend []LegendRow
}

type DayRow struct {
	DayName string
	Cells   []Cell
}

type Cell struct {
	Title    string
	Stripes  []Stripe
	Inactive bool
	InRange  bool
}

type Stripe struct {
	Class string
}

type LegendRow struct {
	Label    string
	Swatches []string
}
