package topartists

type Artist struct {
	Name        string
	TrackCount  int
	AlbumCount  int
	PlayCount   int
	FirstPlayed string
	ImagePath   string
}

type Data struct {
	Artists []Artist
}
