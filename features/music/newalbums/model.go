package newalbums

// Album is one cached entry in the new albums grid.
type Album struct {
	Name       string `json:"name"`
	ArtistName string `json:"artist_name"`
	ImagePath  string `json:"image_path"`
}

// Data is the view model for the newalbums shortcode.
type Data struct {
	Albums []Album
}
