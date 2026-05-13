package char

type Data struct {
	Name              string
	Realm             string
	Level             int
	Spec              string
	Class             string
	Race              string
	Gender            string
	EquippedItemLevel string
	ImagePath         string
	CSSClass          string
	RealmLower        string
	NameLower         string
	Professions       []Profession
	MythicPlus        *MythicPlusData
}

type Profession struct {
	Name       string
	LatestTier ProfessionTier
}

type ProfessionTier struct {
	TierID         int
	Name           string
	SkillPoints    int
	MaxSkillPoints int
}

type MythicPlusRun struct {
	DungeonName   string
	KeystoneLevel int
	Duration      string
	Overtime      bool
	Rating        string
}

type MythicPlusData struct {
	Runs        []MythicPlusRun
	TotalRating string
}
