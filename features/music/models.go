package music

import "time"

type musicArtist struct {
	ID       int    `db:"id"`
	SubsonicID string `db:"subsonic_id"`
	Name     string `db:"name"`
	SortName string `db:"sort_name"`
}

type musicAlbum struct {
	ID       int    `db:"id"`
	SubsonicID string `db:"subsonic_id"`
	Name     string `db:"name"`
	SortName string `db:"sort_name"`
	Year     *int   `db:"year"`
	ArtistID *int   `db:"artist_id"`
}

type musicTrack struct {
	ID          int    `db:"id"`
	SubsonicID  string `db:"subsonic_id"`
	AlbumID     int    `db:"album_id"`
	Name        string `db:"name"`
	Duration    *int   `db:"duration"`
	DiscNumber  *int   `db:"disc_number"`
	TrackNumber *int   `db:"track_number"`
}

type musicPlay struct {
	ID        int       `db:"id"`
	TrackID   int       `db:"track_id"`
	PlayedAt  time.Time `db:"played_at"`
	PlayCount int       `db:"play_count"`
}
