package blizzard

type CharacterProfile struct {
	Name               string       `json:"name"`
	Gender             TypedName    `json:"gender"`
	Faction            TypedName    `json:"faction"`
	Race               NamedRef     `json:"race"`
	CharacterClass     NamedRef     `json:"character_class"`
	ActiveSpec         NamedRef     `json:"active_spec"`
	Realm              RealmRef     `json:"realm"`
	Guild              *GuildRef    `json:"guild"`
	Level              int          `json:"level"`
	LastLoginTimestamp int64        `json:"last_login_timestamp"`
	EquippedItemLevel  int          `json:"equipped_item_level"`
	ActiveTitle        *ActiveTitle `json:"active_title"`
}

type RealmRef struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
	Slug string `json:"slug"`
}

type GuildRef struct {
	Name  string   `json:"name"`
	ID    int      `json:"id"`
	Realm RealmRef `json:"realm"`
}

type ActiveTitle struct {
	Name          string `json:"name"`
	DisplayString string `json:"display_string"`
	ID            int    `json:"id"`
}

type NamedRef struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type TypedName struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type CharacterMedia struct {
	Assets []MediaAsset `json:"assets"`
}

type MediaAsset struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CharacterProfessions struct {
	Primaries   []ProfessionEntry `json:"primaries"`
	Secondaries []ProfessionEntry `json:"secondaries"`
}

type ProfessionEntry struct {
	Profession NamedRef         `json:"profession"`
	Tiers      []ProfessionTier `json:"tiers"`
}

type ProfessionTier struct {
	SkillPoints    int      `json:"skill_points"`
	MaxSkillPoints int      `json:"max_skill_points"`
	Tier           NamedRef `json:"tier"`
}

type MythicKeystoneProfile struct {
	Seasons []NamedRef `json:"seasons"`
}

type MythicKeystoneSeasonProfile struct {
	Season    NamedRef   `json:"season"`
	BestRuns  []MythicRun `json:"best_runs"`
}

type MythicRun struct {
	CompletedTimestamp    int64    `json:"completed_timestamp"`
	Duration              int64    `json:"duration"`
	KeystoneLevel         int      `json:"keystone_level"`
	Dungeon               NamedRef `json:"dungeon"`
	IsCompletedWithinTime bool     `json:"is_completed_within_time"`
	MythicRating          *struct {
		Rating float64 `json:"rating"`
	} `json:"mythic_rating"`
}
