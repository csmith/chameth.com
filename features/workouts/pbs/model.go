package pbs

// Data is the view model for the workoutpbs shortcode.
type Data struct {
	Title string
	Rows  []Row
}

// Row is one table row: the current record for a single segment distance.
type Row struct {
	Distance string
	Time     string
	Pace     string
	Date     string
}
