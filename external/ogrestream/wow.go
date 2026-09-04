package ogrestream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Character is a tracked WoW character with the newest snapshot of each
// dataset as of the requested moment. A section is omitted from the
// response when there is no data of that kind yet.
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

// Portrait is the character's render. Portraits are immutable and
// cacheable forever, identified by the SHA-256 of their bytes; Path is
// where the image can be fetched from.
type Portrait struct {
	Sha256 string `json:"sha256"`
	Path   string `json:"path"`
}

// Achievement is one achievement completion for the account. A completion
// shared by several characters appears once, at its earliest completion.
type Achievement struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	CompletedAt time.Time `json:"completed_at"`
}

// Character fetches a tracked character, optionally as of a past moment:
// at is an RFC3339 timestamp or a bare YYYY-MM-DD (empty means now).
func (c *Client) Character(ctx context.Context, realm, name, at string) (*Character, error) {
	query := url.Values{}
	if at != "" {
		query.Set("at", at)
	}

	path := "/api/wow/characters/" + url.PathEscape(realm) + "/" + url.PathEscape(name)
	b, _, err := c.get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var character Character
	if err := json.Unmarshal(b, &character); err != nil {
		return nil, fmt.Errorf("failed to decode character: %w", err)
	}
	return &character, nil
}

// Achievements fetches the account's recent achievement completions,
// newest first, across every tracked character.
func (c *Client) Achievements(ctx context.Context, limit int) ([]Achievement, error) {
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}

	b, _, err := c.get(ctx, "/api/wow/achievements", query)
	if err != nil {
		return nil, err
	}

	var res struct {
		Achievements []Achievement `json:"achievements"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return nil, fmt.Errorf("failed to decode achievements: %w", err)
	}
	return res.Achievements, nil
}

// Image fetches an asset served by the API — e.g. a portrait's path — as-is.
func (c *Client) Image(ctx context.Context, path string) ([]byte, string, error) {
	b, headers, err := c.get(ctx, path, nil)
	if err != nil {
		return nil, "", err
	}
	return b, headers.Get("Content-Type"), nil
}
