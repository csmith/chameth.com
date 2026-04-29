package wowchar

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
