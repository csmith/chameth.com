package wow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"chameth.com/chameth.com/external/twowords"
)

var ogreStream = twowords.Service{Name: "Ogre Stream", BaseURL: "https://os.yak-wall.ts.net"}

type Character struct {
	Game        string       `json:"game"`
	Realm       string       `json:"realm"`
	Name        string       `json:"name"`
	BlizzardID  int          `json:"blizzard_id"`
	Profile     *Profile     `json:"profile"`
	Professions *Professions `json:"professions"`
	MythicPlus  *MythicPlus  `json:"mythic_plus"`
	Portrait    *Portrait    `json:"portrait"`
}

type Profile struct {
	CapturedAt        time.Time `json:"captured_at"`
	Name              string    `json:"name"`
	Realm             string    `json:"realm"`
	Gender            string    `json:"gender"`
	Faction           string    `json:"faction"`
	Race              string    `json:"race"`
	Class             string    `json:"class"`
	Spec              *string   `json:"spec"`
	Guild             *string   `json:"guild"`
	Title             *string   `json:"title"`
	Level             int       `json:"level"`
	EquippedItemLevel int       `json:"equipped_item_level"`
	LastLogin         time.Time `json:"last_login"`
}

type Professions struct {
	CapturedAt  time.Time    `json:"captured_at"`
	Professions []Profession `json:"professions"`
}

type Profession struct {
	Kind           string `json:"kind"`
	ProfessionID   int    `json:"profession_id"`
	ProfessionName string `json:"profession_name"`
	TierID         int    `json:"tier_id"`
	TierName       string `json:"tier_name"`
	SkillPoints    int    `json:"skill_points"`
	MaxSkillPoints int    `json:"max_skill_points"`
}

type MythicPlus struct {
	CapturedAt  time.Time   `json:"captured_at"`
	SeasonID    int         `json:"season_id"`
	TotalRating float64     `json:"total_rating"`
	Runs        []MythicRun `json:"runs"`
}

type MythicRun struct {
	DungeonID     int       `json:"dungeon_id"`
	DungeonName   string    `json:"dungeon_name"`
	KeystoneLevel int       `json:"keystone_level"`
	CompletedAt   time.Time `json:"completed_at"`
	DurationMS    int64     `json:"duration_ms"`
	InTime        bool      `json:"in_time"`
	Rating        float64   `json:"rating"`
}

type Portrait struct {
	Sha256 string `json:"sha256"`
	Path   string `json:"path"`
}

type Achievement struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	CompletedAt time.Time `json:"completed_at"`
}

func GetCharacter(ctx context.Context, client *http.Client, realm, name, at string) (*Character, error) {
	body, _, err := ogreStream.Get(
		ctx,
		client,
		twowords.Path("/api/wow/characters/%s/%s", realm, name),
		twowords.Query{"at": at},
	)
	if err != nil {
		return nil, err
	}

	var character Character
	if err := json.Unmarshal(body, &character); err != nil {
		return nil, fmt.Errorf("failed to decode character: %w", err)
	}
	return &character, nil
}

func RecentAchievements(ctx context.Context, client *http.Client, limit int) ([]Achievement, error) {
	body, _, err := ogreStream.Get(ctx, client, "/api/wow/achievements", twowords.Query{"limit": limit})
	if err != nil {
		return nil, err
	}

	var res struct {
		Achievements []Achievement `json:"achievements"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode achievements: %w", err)
	}
	return res.Achievements, nil
}

func FetchImage(ctx context.Context, client *http.Client, path string) ([]byte, string, error) {
	body, headers, err := ogreStream.Get(ctx, client, path, nil)
	if err != nil {
		return nil, "", err
	}
	return body, headers.Get("Content-Type"), nil
}
