package nowplaying

import "time"

type nowPlaying struct {
	ArtistName string    `db:"artist_name"`
	TrackName  string    `db:"track_name"`
	AlbumName  string    `db:"album_name"`
	ImagePath  *string   `db:"image_path"`
	PlayedAt   time.Time `db:"played_at"`
}

type Data struct {
	ArtistName string
	TrackName  string
	AlbumName  string
	ImagePath  string
	Status     string
}
