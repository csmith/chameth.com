package wow

import (
	"context"
	"fmt"
	"time"

	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/external/blizzard"
	"chameth.com/chameth.com/features/media"

	"github.com/lib/pq"
)

func AllCharacters(ctx context.Context) ([]Character, error) {
	characters, err := db.Select[Character](ctx, `SELECT * FROM wow_characters ORDER BY character_name`)
	if err != nil {
		return nil, fmt.Errorf("failed to get characters: %w", err)
	}
	return characters, nil
}

func GetCharacter(ctx context.Context, realm, name string) (*Character, error) {
	c, err := db.Get[Character](ctx, `
		SELECT * FROM wow_characters WHERE realm_name = $1 AND character_name = $2
	`, realm, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}
	return &c, nil
}

func getCharacterByID(ctx context.Context, id int) (*Character, error) {
	c, err := db.Get[Character](ctx, `SELECT * FROM wow_characters WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}
	return &c, nil
}

func staleCharacters(ctx context.Context) ([]Character, error) {
	characters, err := db.Select[Character](ctx, `
		SELECT * FROM wow_characters
		WHERE updated_at < NOW() - INTERVAL '24 hours'
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get stale characters: %w", err)
	}
	return characters, nil
}

func upsertCharacter(ctx context.Context, realmName string, profile *blizzard.CharacterProfile) (int, error) {
	var guildName *string
	if profile.Guild != nil {
		guildName = &profile.Guild.Name
	}

	lastLogin := time.UnixMilli(profile.LastLoginTimestamp)

	var equippedLevel *int
	if profile.EquippedItemLevel > 0 {
		equippedLevel = &profile.EquippedItemLevel
	}

	var title *string
	if profile.ActiveTitle != nil {
		title = &profile.ActiveTitle.DisplayString
	}

	id, err := db.Get[int](ctx, `
		INSERT INTO wow_characters (character_name, realm_name, race, class, spec, gender, faction, guild_name, last_login, equipped_item_level, title, level, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (realm_name, character_name)
		DO UPDATE SET
			race = EXCLUDED.race,
			class = EXCLUDED.class,
			spec = EXCLUDED.spec,
			gender = EXCLUDED.gender,
			faction = EXCLUDED.faction,
			guild_name = EXCLUDED.guild_name,
			last_login = EXCLUDED.last_login,
			equipped_item_level = EXCLUDED.equipped_item_level,
			title = EXCLUDED.title,
			level = EXCLUDED.level,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`, profile.Name, realmName, profile.Race.Name,
		profile.CharacterClass.Name, profile.ActiveSpec.Name,
		profile.Gender.Name, profile.Faction.Name,
		guildName, lastLogin, equippedLevel, title, profile.Level)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert character: %w", err)
	}
	return id, nil
}

func saveCharacterImage(ctx context.Context, characterID int, name string, imgData []byte, width, height int) error {
	filename := fmt.Sprintf("%s.png", name)
	mediaPath := fmt.Sprintf("/wow/characters/%s.png", name)

	existing, err := media.GetMediaRelationsForEntity(ctx, "wow_character", characterID)
	if err != nil {
		return fmt.Errorf("failed to check existing media: %w", err)
	}

	for _, rel := range existing {
		if rel.Role != nil && *rel.Role == "render" {
			if err := media.UpdateMedia(ctx, rel.MediaID, "image/png", filename, imgData, &width, &height); err != nil {
				return fmt.Errorf("failed to update media: %w", err)
			}
			return nil
		}
	}

	mediaID, err := media.CreateMedia(ctx, "image/png", filename, imgData, &width, &height, nil)
	if err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}

	caption := name
	role := "render"
	if err := media.CreateMediaRelation(ctx, "wow_character", characterID, mediaID, mediaPath, &caption, nil, &role); err != nil {
		return fmt.Errorf("failed to create media relation: %w", err)
	}

	return nil
}

func upsertMythicPlusRun(ctx context.Context, characterID, seasonID int, run *blizzard.MythicRun) error {
	rating := 0.0
	if run.MythicRating != nil {
		rating = run.MythicRating.Rating
	}

	_, err := db.Exec(ctx, `
		INSERT INTO wow_mythic_runs (character_id, season_id, dungeon_id, dungeon_name, completed_timestamp, duration, keystone_level, is_completed_within_time, mythic_rating)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (character_id, season_id, dungeon_id)
		DO UPDATE SET
			dungeon_name = EXCLUDED.dungeon_name,
			completed_timestamp = EXCLUDED.completed_timestamp,
			duration = EXCLUDED.duration,
			keystone_level = EXCLUDED.keystone_level,
			is_completed_within_time = EXCLUDED.is_completed_within_time,
			mythic_rating = EXCLUDED.mythic_rating
		WHERE EXCLUDED.mythic_rating >= wow_mythic_runs.mythic_rating
	`, characterID, seasonID, run.Dungeon.ID, run.Dungeon.Name, run.CompletedTimestamp, run.Duration, run.KeystoneLevel, run.IsCompletedWithinTime, rating)
	if err != nil {
		return fmt.Errorf("failed to upsert mythic+ run: %w", err)
	}
	return nil
}

func GetCharacterMythicPlusRuns(ctx context.Context, characterID, seasonID int) ([]MythicPlusRun, error) {
	runs, err := db.Select[MythicPlusRun](ctx, `
		SELECT * FROM wow_mythic_runs WHERE character_id = $1 AND season_id = $2 ORDER BY dungeon_name
	`, characterID, seasonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mythic+ runs: %w", err)
	}
	return runs, nil
}

func GetCurrentSeasonID(ctx context.Context) (int, error) {
	id, err := db.Get[int](ctx, `SELECT MAX(season_id) FROM wow_mythic_runs`)
	if err != nil {
		return 0, fmt.Errorf("failed to get current season id: %w", err)
	}
	return id, nil
}

func GetCharacterProfessions(ctx context.Context, characterID int) ([]CharacterProfession, error) {
	professions, err := db.Select[CharacterProfession](ctx, `
		SELECT * FROM wow_character_professions WHERE character_id = $1 ORDER BY kind, profession_name, tier_name
	`, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character professions: %w", err)
	}
	return professions, nil
}

func syncProfessions(ctx context.Context, characterID int, profs *blizzard.CharacterProfessions) error {
	var allTierIDs []int

	for _, p := range profs.Primaries {
		for _, t := range p.Tiers {
			allTierIDs = append(allTierIDs, t.Tier.ID)
		}
	}
	for _, p := range profs.Secondaries {
		if p.Profession.ID == 794 {
			continue
		}
		for _, t := range p.Tiers {
			allTierIDs = append(allTierIDs, t.Tier.ID)
		}
	}

	if len(allTierIDs) > 0 {
		_, err := db.Exec(ctx, `
			DELETE FROM wow_character_professions
			WHERE character_id = $1 AND NOT (tier_id = ANY($2))
		`, characterID, pq.Array(allTierIDs))
		if err != nil {
			return fmt.Errorf("failed to delete old professions: %w", err)
		}
	} else {
		_, err := db.Exec(ctx, `DELETE FROM wow_character_professions WHERE character_id = $1`, characterID)
		if err != nil {
			return fmt.Errorf("failed to delete all professions: %w", err)
		}
	}

	for _, p := range profs.Primaries {
		for _, t := range p.Tiers {
			_, err := db.Exec(ctx, `
				INSERT INTO wow_character_professions (character_id, tier_id, tier_name, profession_id, profession_name, skill_points, max_skill_points, kind)
				VALUES ($1, $2, $3, $4, $5, $6, $7, 'primary')
				ON CONFLICT (character_id, tier_id)
				DO UPDATE SET
					tier_name = EXCLUDED.tier_name,
					profession_id = EXCLUDED.profession_id,
					profession_name = EXCLUDED.profession_name,
					skill_points = EXCLUDED.skill_points,
					max_skill_points = EXCLUDED.max_skill_points,
					kind = EXCLUDED.kind
			`, characterID, t.Tier.ID, t.Tier.Name, p.Profession.ID, p.Profession.Name, t.SkillPoints, t.MaxSkillPoints)
			if err != nil {
				return fmt.Errorf("failed to upsert profession: %w", err)
			}
		}
	}

	for _, p := range profs.Secondaries {
		if p.Profession.ID == 794 {
			continue
		}
		for _, t := range p.Tiers {
			_, err := db.Exec(ctx, `
				INSERT INTO wow_character_professions (character_id, tier_id, tier_name, profession_id, profession_name, skill_points, max_skill_points, kind)
				VALUES ($1, $2, $3, $4, $5, $6, $7, 'secondary')
				ON CONFLICT (character_id, tier_id)
				DO UPDATE SET
					tier_name = EXCLUDED.tier_name,
					profession_id = EXCLUDED.profession_id,
					profession_name = EXCLUDED.profession_name,
					skill_points = EXCLUDED.skill_points,
					max_skill_points = EXCLUDED.max_skill_points,
					kind = EXCLUDED.kind
			`, characterID, t.Tier.ID, t.Tier.Name, p.Profession.ID, p.Profession.Name, t.SkillPoints, t.MaxSkillPoints)
			if err != nil {
				return fmt.Errorf("failed to upsert profession: %w", err)
			}
		}
	}

	return nil
}

type RecentAchievement struct {
	AchievementID   int       `db:"achievement_id"`
	AchievementName string    `db:"achievement_name"`
	CompletedAt     time.Time `db:"completed_at"`
	CharacterName   string    `db:"character_name"`
	IsAccountWide   bool      `db:"is_account_wide"`
}

func GetRecentAchievements(ctx context.Context, limit int) ([]RecentAchievement, error) {
	achievements, err := db.Select[RecentAchievement](ctx, `
		WITH ordered AS (
			SELECT
				a.achievement_id,
				a.achievement_name,
				a.completed_at,
				a.character_id,
				LAG(a.completed_at) OVER (PARTITION BY a.achievement_id ORDER BY a.completed_at) AS prev_at
			FROM wow_achievements a
		),
		islands AS (
			SELECT
				achievement_id,
				achievement_name,
				completed_at,
				character_id,
				SUM(CASE WHEN prev_at IS NULL OR EXTRACT(EPOCH FROM (completed_at - prev_at)) >= 60 THEN 1 ELSE 0 END)
					OVER (PARTITION BY achievement_id ORDER BY completed_at) AS grp
			FROM ordered
		),
		deduped AS (
			SELECT DISTINCT ON (achievement_id, grp)
				achievement_id, achievement_name, completed_at, character_id, grp
			FROM islands
			ORDER BY achievement_id, grp, completed_at
		)
		SELECT
			d.achievement_id,
			d.achievement_name,
			d.completed_at,
			c.character_name,
			(SELECT COUNT(*) > 1 FROM islands i2 WHERE i2.achievement_id = d.achievement_id AND i2.grp = d.grp) AS is_account_wide
		FROM deduped d
		JOIN wow_characters c ON c.id = d.character_id
		ORDER BY d.completed_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent achievements: %w", err)
	}
	return achievements, nil
}

func syncAchievements(ctx context.Context, characterID int, achievements *blizzard.CharacterAchievements) error {
	for _, a := range achievements.Achievements {
		if a.CompletedTimestamp == 0 {
			continue
		}

		completedAt := time.UnixMilli(a.CompletedTimestamp)

		_, err := db.Exec(ctx, `
			INSERT INTO wow_achievements (achievement_id, achievement_name, completed_at, character_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (achievement_id, completed_at, character_id)
			DO NOTHING
		`, a.Achievement.ID, a.Achievement.Name, completedAt, characterID)
		if err != nil {
			return fmt.Errorf("failed to upsert achievement %d: %w", a.Achievement.ID, err)
		}
	}

	return nil
}
