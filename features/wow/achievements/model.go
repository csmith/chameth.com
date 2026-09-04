package achievements

import "time"

// Data is the cached account achievements as rendered by the shortcode.
type Data struct {
	Achievements []Achievement `json:"achievements"`
}

// Achievement is one achievement completion for the account; the service
// provides no per-character attribution.
type Achievement struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	CompletedAt time.Time `json:"completed_at"`
}
