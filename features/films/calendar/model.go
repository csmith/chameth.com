package calendar

type Data struct {
	Title string
	Rows  []DayRow
}

type DayRow struct {
	DayName string
	Cells   []Cell
}

type Cell struct {
	Title    string
	Class    string
	Inactive bool
}
