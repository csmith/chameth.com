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

type CharacterProfession struct {
	CharacterID    int    `db:"character_id"`
	TierID         int    `db:"tier_id"`
	TierName       string `db:"tier_name"`
	ProfessionID   int    `db:"profession_id"`
	ProfessionName string `db:"profession_name"`
	SkillPoints    int    `db:"skill_points"`
	MaxSkillPoints int    `db:"max_skill_points"`
	Kind           string `db:"kind"`
}

type MythicPlusRun struct {
	ID                    int     `db:"id"`
	CharacterID           int     `db:"character_id"`
	SeasonID              int     `db:"season_id"`
	DungeonID             int     `db:"dungeon_id"`
	DungeonName           string  `db:"dungeon_name"`
	CompletedTimestamp     int64   `db:"completed_timestamp"`
	Duration              int64   `db:"duration"`
	KeystoneLevel         int     `db:"keystone_level"`
	IsCompletedWithinTime bool    `db:"is_completed_within_time"`
	MythicRating          float64 `db:"mythic_rating"`
}
