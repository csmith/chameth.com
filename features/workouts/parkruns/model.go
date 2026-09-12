package parkruns

type Box struct {
	Title  string
	Value  string
	Detail string
}

type StatsData struct {
	Total  Box
	Venues Box
	Best   Box
}

type TableData struct {
	Title string
	Rows  []TableRow
}

type TableRow struct {
	Date   string
	Event  string
	Time   string
	Pos    string
	AgePos string
}
