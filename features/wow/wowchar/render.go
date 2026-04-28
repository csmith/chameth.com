package wowchar

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"strings"

	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/features/wow"
)

//go:embed *.gotpl
var templates string

var tmpl = template.Must(template.New("wowchar.html.gotpl").Parse(templates))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("wowchar requires 2 arguments (realm character)")
	}

	c, err := wow.GetCharacter(ctx, args[0], args[1])
	if err != nil {
		return "", fmt.Errorf("failed to get character: %w", err)
	}

	itemLevel := ""
	if c.EquippedItemLevel != nil {
		itemLevel = fmt.Sprintf("%d", *c.EquippedItemLevel)
	}

	professions, err := wow.GetCharacterProfessions(ctx, c.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get character professions: %w", err)
	}

	var dataProfessions []Profession
	profMap := make(map[int]*Profession)
	for _, p := range professions {
		prof, ok := profMap[p.ProfessionID]
		if !ok {
			dataProfessions = append(dataProfessions, Profession{Name: p.ProfessionName})
			prof = &dataProfessions[len(dataProfessions)-1]
			profMap[p.ProfessionID] = prof
		}
		tier := ProfessionTier{
			TierID:         p.TierID,
			Name:           p.TierName,
			SkillPoints:    p.SkillPoints,
			MaxSkillPoints: p.MaxSkillPoints,
		}
		if tier.TierID > prof.LatestTier.TierID {
			prof.LatestTier = tier
		}
	}

	return renderTemplate(Data{
		Name:              c.CharacterName,
		Realm:             c.RealmName,
		Level:             c.Level,
		Spec:              c.Spec,
		Class:             c.Class,
		Race:              c.Race,
		Gender:            c.Gender,
		EquippedItemLevel: itemLevel,
		CSSClass:          "wow-class-" + strings.ToLower(strings.ReplaceAll(c.Class, " ", "-")),
		RealmLower:        strings.ToLower(c.RealmName),
		NameLower:         strings.ToLower(c.CharacterName),
		Professions:       dataProfessions,
	})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
