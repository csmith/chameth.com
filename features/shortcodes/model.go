package shortcodes

import "time"

// shortcodeData is one cached row of externally-retrieved shortcode data.
type shortcodeData struct {
	ID            int        `db:"id"`
	Shortcode     string     `db:"shortcode"`
	Version       int        `db:"version"`
	ArgsHash      string     `db:"args_hash"`
	Args          []byte     `db:"args"`
	Data          []byte     `db:"data"` // nil when the last retrieval failed
	NextRefreshAt *time.Time `db:"next_refresh_at"`
	RetrievedAt   time.Time  `db:"retrieved_at"`
}
