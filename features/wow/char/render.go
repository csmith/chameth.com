package char

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"

	"chameth.com/chameth.com/features/wow"
)

//go:embed *.gotpl
var templates string

var tmpl = template.Must(template.New("char.html.gotpl").Parse(templates))

// fishingProfessionID is Blizzard's id for Fishing, whose tiers don't fit
// the skill-points model the professions table renders, so it is skipped
// (as the old sync did).
const fishingProfessionID = 794

// buildData maps the API response onto the cached render model.
func buildData(c *wow.Character, imagePath string) Data {
	p := c.Profile

	data := Data{
		Name:       p.Name,
		Realm:      p.Realm,
		Level:      p.Level,
		Class:      p.Class,
		Race:       p.Race,
		Gender:     p.Gender,
		ImagePath:  imagePath,
		CSSClass:   "wow-class-" + strings.ToLower(strings.ReplaceAll(p.Class, " ", "-")),
		RealmLower: strings.ToLower(p.Realm),
		NameLower:  strings.ToLower(p.Name),
	}
	if p.Spec != nil {
		data.Spec = *p.Spec
	}
	if p.EquippedItemLevel > 0 {
		data.EquippedItemLevel = strconv.Itoa(p.EquippedItemLevel)
	}
	if c.Professions != nil {
		data.Professions = buildProfessions(c.Professions.Professions)
	}
	if c.MythicPlus != nil {
		data.MythicPlus = buildMythicPlus(c.MythicPlus)
	}

	return data
}

// buildProfessions collapses the tiers to each profession's current one,
// primaries before secondaries, each alphabetically.
func buildProfessions(professions []wow.Profession) []Profession {
	var built []Profession
	indexes := make(map[int]int)

	for _, p := range professions {
		if p.ProfessionID == fishingProfessionID {
			continue
		}

		i, ok := indexes[p.ProfessionID]
		if !ok {
			i = len(built)
			indexes[p.ProfessionID] = i
			built = append(built, Profession{Name: p.ProfessionName, Kind: p.Kind})
		}
		if p.TierID > built[i].LatestTier.TierID {
			built[i].LatestTier = ProfessionTier{
				TierID:         p.TierID,
				Name:           p.TierName,
				SkillPoints:    p.SkillPoints,
				MaxSkillPoints: p.MaxSkillPoints,
			}
		}
	}

	sort.SliceStable(built, func(i, j int) bool {
		if built[i].Kind != built[j].Kind {
			return built[i].Kind < built[j].Kind
		}
		return built[i].Name < built[j].Name
	})

	return built
}

// buildMythicPlus maps the season's runs to the render model. A dungeon
// can appear twice — its best timed and best untimed attempt — and only
// the highest-rated one is displayed, ordered by dungeon name as the
// local table used to serve them.
func buildMythicPlus(mp *wow.MythicPlus) *MythicPlusData {
	best := make(map[int]wow.MythicRun, len(mp.Runs))
	for _, r := range mp.Runs {
		current, ok := best[r.DungeonID]
		if !ok || r.Rating > current.Rating || (r.Rating == current.Rating && r.InTime && !current.InTime) {
			best[r.DungeonID] = r
		}
	}

	runs := make([]MythicPlusRun, 0, len(best))
	for _, r := range best {
		runs = append(runs, MythicPlusRun{
			DungeonName:   r.DungeonName,
			KeystoneLevel: r.KeystoneLevel,
			Duration:      formatDuration(r.DurationMS),
			Overtime:      !r.InTime,
			Rating:        fmt.Sprintf("%.0f", r.Rating),
		})
	}
	sort.SliceStable(runs, func(i, j int) bool {
		return runs[i].DungeonName < runs[j].DungeonName
	})

	return &MythicPlusData{
		Runs:        runs,
		TotalRating: fmt.Sprintf("%.0f", mp.TotalRating),
	}
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), err
}

func formatDuration(ms int64) string {
	totalSeconds := ms / 1000
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
