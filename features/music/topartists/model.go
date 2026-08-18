package topartists

type topArtist struct {
	Name       string  `db:"name"`
	TrackCount int     `db:"track_count"`
	AlbumCount int     `db:"album_count"`
	PlayCount  int     `db:"play_count"`
	ImagePath  *string `db:"image_path"`
}

type Artist struct {
	Position   int
	Name       string
	TrackCount int
	AlbumCount int
	PlayCount  int
	ImagePath  string
}

type Data struct {
	Artists []Artist
}
