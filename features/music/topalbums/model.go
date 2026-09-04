package topalbums

// Album is one cached row of the all-time album ranking.
type Album struct {
	Position   int    `json:"position"`
	Name       string `json:"name"`
	ArtistName string `json:"artist_name"`
	TrackCount int    `json:"track_count"`
	PlayCount  int    `json:"play_count"`
	ImagePath  string `json:"image_path"`
}

// Data is the view model for the topalbums shortcode.
type Data struct {
	Albums []Album
}
