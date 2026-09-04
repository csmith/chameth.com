package nowplaying

import "time"

// cached is the persisted now-playing state; the relative status is
// computed at render time from PlayedAt.
type cached struct {
	ArtistName string    `json:"artist_name"`
	TrackName  string    `json:"track_name"`
	AlbumName  string    `json:"album_name"`
	ImagePath  string    `json:"image_path"`
	PlayedAt   time.Time `json:"played_at"`
}

// Data is the view model for the nowplaying shortcode.
type Data struct {
	ArtistName string
	TrackName  string
	AlbumName  string
	ImagePath  string
	Status     string
}
