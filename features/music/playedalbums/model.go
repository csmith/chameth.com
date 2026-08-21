package playedalbums

type playedAlbum struct {
	Position         int     `db:"position"`
	Name             string  `db:"name"`
	ArtistName       string  `db:"artist_name"`
	TrackCount       int     `db:"track_count"`
	PlayCount        int     `db:"play_count"`
	PreviousPosition *int    `db:"previous_position"`
	ImagePath        *string `db:"image_path"`
}

// Data is the view model for the playedalbums shortcode.
type Data struct {
	Title  string
	Albums []Album
}

// Album is one row of the most played albums table.
type Album struct {
	Position      int
	Movement      string // up, down, same or new, versus the previous period
	MovementTitle string // human-readable description for the row tooltip
	Name          string
	ArtistName    string
	TrackCount    int
	PlayCount     int
	ImagePath     string
}
