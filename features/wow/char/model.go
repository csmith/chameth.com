package char

type Data struct {
	Name              string          `json:"name"`
	Realm             string          `json:"realm"`
	Level             int             `json:"level"`
	Spec              string          `json:"spec"`
	Class             string          `json:"class"`
	Race              string          `json:"race"`
	Gender            string          `json:"gender"`
	EquippedItemLevel string          `json:"equipped_item_level"`
	ImagePath         string          `json:"image_path"`
	CSSClass          string          `json:"css_class"`
	RealmLower        string          `json:"realm_lower"`
	NameLower         string          `json:"name_lower"`
	Professions       []Profession    `json:"professions"`
	MythicPlus        *MythicPlusData `json:"mythic_plus"`
}

type Profession struct {
	Name       string         `json:"name"`
	Kind       string         `json:"kind"`
	LatestTier ProfessionTier `json:"latest_tier"`
}

type ProfessionTier struct {
	TierID         int    `json:"tier_id"`
	Name           string `json:"name"`
	SkillPoints    int    `json:"skill_points"`
	MaxSkillPoints int    `json:"max_skill_points"`
}

type MythicPlusRun struct {
	DungeonName   string `json:"dungeon_name"`
	KeystoneLevel int    `json:"keystone_level"`
	Duration      string `json:"duration"`
	Overtime      bool   `json:"overtime"`
	Rating        string `json:"rating"`
}

type MythicPlusData struct {
	Runs        []MythicPlusRun `json:"runs"`
	TotalRating string          `json:"total_rating"`
}
