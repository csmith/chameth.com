package summary

type Stat struct {
	Value string
	Label string
}

type PB struct {
	Label    string
	Time     string
	Previous string
}

type Section struct {
	Title string
	Stats []Stat
	PBs   []PB
}

type Data struct {
	Sections []Section
}
