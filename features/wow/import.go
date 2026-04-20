package wow

import (
	"context"
	"fmt"
	"log/slog"
)

func ImportCharacter(ctx context.Context, realm, name string) error {
	characterID, err := syncCharacter(ctx, realm, name)
	if err != nil {
		return err
	}

	slog.Info("Imported WoW character", "character_id", characterID, "realm", realm, "character", name)
	return nil
}

func RefreshCharacter(ctx context.Context, characterID int) error {
	c, err := getCharacterByID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	if _, err := syncCharacter(ctx, c.RealmName, c.CharacterName); err != nil {
		return err
	}

	slog.Info("Refreshed WoW character", "character_id", characterID, "realm", c.RealmName, "character", c.CharacterName)
	return nil
}

func syncCharacter(ctx context.Context, realm, name string) (int, error) {
	bc := client()

	profile, err := bc.GetCharacterProfile(realm, name)
	if err != nil {
		return 0, fmt.Errorf("failed to get character profile: %w", err)
	}

	characterID, err := upsertCharacter(ctx, realm, profile)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert character: %w", err)
	}

	media, err := bc.GetCharacterMedia(realm, name)
	if err != nil {
		return 0, fmt.Errorf("failed to get character media: %w", err)
	}

	if err := fetchAndSaveCharacterImage(ctx, characterID, profile.Name, media); err != nil {
		return 0, fmt.Errorf("failed to update character image: %w", err)
	}

	return characterID, nil
}
