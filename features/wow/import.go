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

	professions, err := bc.GetCharacterProfessions(realm, name)
	if err != nil {
		return 0, fmt.Errorf("failed to get character professions: %w", err)
	}

	if err := syncProfessions(ctx, characterID, professions); err != nil {
		return 0, fmt.Errorf("failed to sync professions: %w", err)
	}

	mplus, err := bc.GetMythicKeystoneProfile(realm, name)
	if err != nil {
		return 0, fmt.Errorf("failed to get mythic keystone profile: %w", err)
	}

	seasonID := currentSeasonID(mplus)
	if seasonID > 0 {
		season, err := bc.GetMythicKeystoneSeasonProfile(realm, name, seasonID)
		if err != nil {
			return 0, fmt.Errorf("failed to get mythic keystone season profile: %w", err)
		}
		for i := range season.BestRuns {
			if err := upsertMythicPlusRun(ctx, characterID, seasonID, &season.BestRuns[i]); err != nil {
				return 0, fmt.Errorf("failed to upsert mythic+ run: %w", err)
			}
		}
	}

	return characterID, nil
}
