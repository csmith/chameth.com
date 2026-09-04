package playedalbums

// Album is one cached row of the most played albums table.
type Album struct {
	Position      int    `json:"position"`
	Movement      string `json:"movement"`       // up, down, same or new, versus the previous period
	MovementTitle string `json:"movement_title"` // human-readable description for the row tooltip
	Name          string `json:"name"`
	ArtistName    string `json:"artist_name"`
	TrackCount    int    `json:"track_count"`
	PlayCount     int    `json:"play_count"`
	ImagePath     string `json:"image_path"`
}

// Data is the view model for the playedalbums shortcode.
type Data struct {
	Title  string
	Albums []Album
}
