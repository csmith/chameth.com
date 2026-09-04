package wow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const ogreStreamBaseURL = "https://os.yak-wall.ts.net"

var errNotFound = errors.New("not found")

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
	query := url.Values{}
	if at != "" {
		query.Set("at", at)
	}

	path := "/api/wow/characters/" + url.PathEscape(realm) + "/" + url.PathEscape(name)
	body, _, err := ogreStreamGet(ctx, client, path, query)
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
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}

	body, _, err := ogreStreamGet(ctx, client, "/api/wow/achievements", query)
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
	body, headers, err := ogreStreamGet(ctx, client, path, nil)
	if err != nil {
		return nil, "", err
	}
	return body, headers.Get("Content-Type"), nil
}

func ogreStreamGet(ctx context.Context, client *http.Client, path string, query url.Values) ([]byte, http.Header, error) {
	u := strings.TrimRight(ogreStreamBaseURL, "/") + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build request: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch %s: %w", path, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %s response: %w", path, err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, nil, errNotFound
	}
	if res.StatusCode >= 400 {
		slog.Warn("Ogre Stream request failed", "path", path, "status", res.StatusCode, "response", string(body))
		return nil, nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}
	return body, res.Header, nil
}
