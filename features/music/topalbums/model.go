package topalbums

type Album struct {
	Position   int
	Name       string
	ArtistName string
	TrackCount int
	PlayCount  int
	ImagePath  string
}

type Data struct {
	Albums []Album
}
