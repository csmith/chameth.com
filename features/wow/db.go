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
