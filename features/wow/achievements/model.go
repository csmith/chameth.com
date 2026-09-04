package achievements

import "time"

type Data struct {
	Achievements []Achievement `json:"achievements"`
}

type Achievement struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	CompletedAt time.Time `json:"completed_at"`
}
