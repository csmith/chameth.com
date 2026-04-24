package wow

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"chameth.com/chameth.com/external/blizzard"
)

var (
	blizzardClientID     = flag.String("blizzard-client-id", "", "Blizzard API client ID")
	blizzardClientSecret = flag.String("blizzard-client-secret", "", "Blizzard API client secret")

	blizzardOnce   sync.Once
	blizzardClient *blizzard.Client
)

func client() *blizzard.Client {
	blizzardOnce.Do(func() {
		blizzardClient = blizzard.NewClient(&http.Client{Timeout: 30 * time.Second}, *blizzardClientID, *blizzardClientSecret)
	})
	return blizzardClient
}

func RunSync(ctx context.Context) {
	if *blizzardClientID == "" {
		return
	}

	ticker := time.NewTicker(4 * time.Hour)
	defer ticker.Stop()

	syncCharacters(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			syncCharacters(ctx)
		}
	}
}

func syncCharacters(ctx context.Context) {
	characters, err := staleCharacters(ctx)
	if err != nil {
		slog.Error("Failed to get stale WoW characters", "error", err)
		return
	}

	if len(characters) == 0 {
		return
	}

	for _, c := range characters {
		profile, err := client().GetCharacterProfile(c.RealmName, c.CharacterName)
		if err != nil {
			slog.Error("Failed to get character profile", "error", err, "realm", c.RealmName, "character", c.CharacterName)
			continue
		}

		characterID, err := upsertCharacter(ctx, c.RealmName, profile)
		if err != nil {
			slog.Error("Failed to upsert character", "error", err, "character", profile.Name)
			continue
		}

		media, err := client().GetCharacterMedia(c.RealmName, c.CharacterName)
		if err != nil {
			slog.Error("Failed to get character media", "error", err, "realm", c.RealmName, "character", c.CharacterName)
		} else if err := fetchAndSaveCharacterImage(ctx, characterID, profile.Name, media); err != nil {
			slog.Error("Failed to update character image", "error", err, "character", profile.Name)
		}

		professions, err := client().GetCharacterProfessions(c.RealmName, c.CharacterName)
		if err != nil {
			slog.Error("Failed to get character professions", "error", err, "realm", c.RealmName, "character", c.CharacterName)
		} else if err := syncProfessions(ctx, characterID, professions); err != nil {
			slog.Error("Failed to sync professions", "error", err, "character", profile.Name)
		}

		slog.Info("Updated WoW character", "character", profile.Name, "realm", profile.Realm.Name)
	}
}
