package wow

import (
	"context"
	"fmt"
	"time"

	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/external/blizzard"
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

	existing, err := db.GetMediaRelationsForEntity(ctx, "wow_character", characterID)
	if err != nil {
		return fmt.Errorf("failed to check existing media: %w", err)
	}

	for _, rel := range existing {
		if rel.Role != nil && *rel.Role == "render" {
			if err := db.UpdateMedia(ctx, rel.MediaID, "image/png", filename, imgData, &width, &height); err != nil {
				return fmt.Errorf("failed to update media: %w", err)
			}
			return nil
		}
	}

	mediaID, err := db.CreateMedia(ctx, "image/png", filename, imgData, &width, &height, nil)
	if err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}

	caption := name
	role := "render"
	if err := db.CreateMediaRelation(ctx, "wow_character", characterID, mediaID, mediaPath, &caption, nil, &role); err != nil {
		return fmt.Errorf("failed to create media relation: %w", err)
	}

	return nil
}
