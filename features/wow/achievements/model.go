package achievements

import "time"

type Data struct {
	Achievements []Achievement
}

type Achievement struct {
	ID            int
	Name          string
	CompletedAt   time.Time
	CharacterName string
	IsAccountWide bool
}
