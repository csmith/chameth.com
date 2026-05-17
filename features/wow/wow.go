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

func currentSeasonID(p *blizzard.MythicKeystoneProfile) int {
	if len(p.Seasons) == 0 {
		return 0
	}
	maxID := p.Seasons[0].ID
	for _, s := range p.Seasons[1:] {
		if s.ID > maxID {
			maxID = s.ID
		}
	}
	return maxID
}

func RegisterGoroutine(ctx context.Context) func() {
	return func() {
		runSync(ctx)
	}
}

func runSync(ctx context.Context) {
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
		if _, err := syncCharacter(ctx, c.RealmName, c.CharacterName); err != nil {
			slog.Error("Failed to sync character", "error", err, "realm", c.RealmName, "character", c.CharacterName)
		} else {
			slog.Info("Updated WoW character", "realm", c.RealmName, "character", c.CharacterName)
		}
	}
}
