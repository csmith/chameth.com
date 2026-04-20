package wow

import "time"

type Character struct {
	ID                int        `db:"id"`
	CharacterName     string     `db:"character_name"`
	RealmName         string     `db:"realm_name"`
	Race              string     `db:"race"`
	Class             string     `db:"class"`
	Spec              string     `db:"spec"`
	Gender            string     `db:"gender"`
	Faction           string     `db:"faction"`
	GuildName         *string    `db:"guild_name"`
	LastLogin         *time.Time `db:"last_login"`
	EquippedItemLevel *int       `db:"equipped_item_level"`
	Title             *string    `db:"title"`
	Level             int        `db:"level"`
	UpdatedAt         time.Time  `db:"updated_at"`
}
