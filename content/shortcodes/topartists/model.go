package topartists

type Artist struct {
	Position   int
	Name       string
	TrackCount int
	AlbumCount  int
	PlayCount int
	ImagePath string
}

type Data struct {
	Artists []Artist
}
