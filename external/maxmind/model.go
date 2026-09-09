package maxmind

import "time"

type ASNRefresh struct {
	RefreshedAt time.Time `db:"refreshed_at"`
}
