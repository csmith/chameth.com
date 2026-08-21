package newalbums

type newAlbum struct {
	Name       string  `db:"name"`
	ArtistName string  `db:"artist_name"`
	PlayCount  int     `db:"play_count"`
	ImagePath  *string `db:"image_path"`
}

// Data is the view model for the newalbums shortcode.
type Data struct {
	Albums []Album
}

// Album is one entry in the new albums grid.
type Album struct {
	Name       string
	ArtistName string
	ImagePath  string
}
