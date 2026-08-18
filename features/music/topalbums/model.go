package topalbums

type topAlbum struct {
	Name       string  `db:"name"`
	ArtistName string  `db:"artist_name"`
	TrackCount int     `db:"track_count"`
	PlayCount  int     `db:"play_count"`
	ImagePath  *string `db:"image_path"`
}

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
