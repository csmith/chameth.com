package pbs

type Data struct {
	Title string
	Rows  []Row
}

type Row struct {
	Distance string
	Time     string
	Pace     string
	Date     string
}
