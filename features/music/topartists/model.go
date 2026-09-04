package topartists

// Artist is one cached row of the all-time artist ranking.
type Artist struct {
	Position   int    `json:"position"`
	Name       string `json:"name"`
	TrackCount int    `json:"track_count"`
	AlbumCount int    `json:"album_count"`
	PlayCount  int    `json:"play_count"`
	ImagePath  string `json:"image_path"`
}

// Data is the view model for the topartists shortcode.
type Data struct {
	Artists []Artist
}
