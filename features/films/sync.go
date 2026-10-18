package films

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"chameth.com/chameth.com/features/media"
	"chameth.com/chameth.com/features/shortcodes"
	"tailscale.com/tsnet"
)

const syncInterval = time.Hour

// RegisterGoroutine keeps the local mirror of Tired Scroll's index (films,
// lists and list entries) and the rehosted posters up to date.
func RegisterGoroutine(ctx context.Context, mgr *shortcodes.Manager, ts *tsnet.Server) func() {
	return func() {
		client := ts.HTTPClient()
		runSync(ctx, mgr, client)

		ticker := time.NewTicker(syncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runSync(ctx, mgr, client)
			}
		}
	}
}

func runSync(ctx context.Context, mgr *shortcodes.Manager, client *http.Client) {
	if err := syncIndex(ctx, mgr, client); err != nil {
		slog.Error("Failed to sync films from Tired Scroll", "error", err)
	}
}

func syncIndex(ctx context.Context, mgr *shortcodes.Manager, client *http.Client) error {
	idx, err := fetchIndex(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to fetch index: %w", err)
	}
	if len(idx.Films) == 0 {
		// Applying this would delete every film and poster.
		return errors.New("index contains no films")
	}

	states, err := getFilmSyncStates(ctx)
	if err != nil {
		return fmt.Errorf("failed to load film states: %w", err)
	}
	previous := make(map[int]filmSyncState, len(states))
	for _, s := range states {
		previous[s.ID] = s
	}

	entries, err := changedEntries(ctx, idx)
	if err != nil {
		return err
	}

	if err := replaceIndex(ctx, idx, entries); err != nil {
		return err
	}

	for _, f := range idx.Films {
		prev, known := previous[f.ID]

		if f.Poster != nil && (!known || prev.PosterUpdatedAt == nil || !prev.PosterUpdatedAt.Equal(f.Poster.UpdatedAt)) {
			if err := syncPoster(ctx, client, f.FilmSummary); err != nil {
				slog.Error("Failed to rehost film poster", "film_id", f.ID, "error", err)
			}
		}

		if !known || prev.Path != FilmPath(f.Slug) || formatDate(prev.LastWatched) != derefString(f.LastWatched) {
			// A new watch (or a film new to the site) should show up without
			// waiting for the page's scheduled refresh.
			if err := mgr.RefreshData(ctx, filmPageData, []string{f.Slug}); err != nil {
				slog.Error("Failed to refresh film page data", "film_id", f.ID, "error", err)
			}
		}
	}

	if err := removeOrphanedPosters(ctx); err != nil {
		slog.Error("Failed to remove posters of removed films", "error", err)
	}

	return nil
}

// changedEntries returns the list entries described by the index, or nil if
// they match what is already stored.
func changedEntries(ctx context.Context, idx *index) ([]listEntryRow, error) {
	lists := slices.Clone(idx.Lists)
	slices.SortFunc(lists, func(a, b indexList) int { return a.ID - b.ID })

	entries := []listEntryRow{}
	for _, l := range lists {
		for _, e := range l.Entries {
			entries = append(entries, listEntryRow{ListID: l.ID, FilmID: e.FilmID, Position: e.Position})
		}
	}

	existing, err := getListEntryRows(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load film list entries: %w", err)
	}
	if slices.Equal(existing, entries) {
		return nil, nil
	}
	return entries, nil
}

func syncPoster(ctx context.Context, client *http.Client, f FilmSummary) error {
	data, contentType, err := fetchPoster(ctx, client, f.Poster.Path)
	if err != nil {
		return fmt.Errorf("failed to download poster: %w", err)
	}

	path := posterPath(f.ID, contentType)
	filename := fmt.Sprintf("%d.%s", f.ID, posterExtension(contentType))
	width, height := imageDimensions(data)

	existing, err := getPosterRelations(ctx, f.ID)
	if err != nil {
		return fmt.Errorf("failed to look up existing poster: %w", err)
	}

	stored := false
	for _, rel := range existing {
		if rel.Path == path {
			if err := media.UpdateMedia(ctx, rel.MediaID, contentType, filename, data, width, height); err != nil {
				return fmt.Errorf("failed to update poster: %w", err)
			}
			stored = true
		}
	}

	if !stored {
		mediaID, err := media.CreateMedia(ctx, contentType, filename, data, width, height, nil)
		if err != nil {
			return fmt.Errorf("failed to create poster media: %w", err)
		}
		caption := f.Title
		description := "Poster of " + f.Title
		role := "poster"
		if err := media.CreateMediaRelation(ctx, "film", f.ID, mediaID, path, &caption, &description, &role); err != nil {
			return fmt.Errorf("failed to create poster relation: %w", err)
		}
	}

	// A poster whose type changed lives at a different path.
	for _, rel := range existing {
		if rel.Path != path {
			if err := deletePoster(ctx, rel); err != nil {
				return err
			}
		}
	}

	return setPosterUpdatedAt(ctx, f.ID, f.Poster.UpdatedAt)
}

func removeOrphanedPosters(ctx context.Context) error {
	orphans, err := getOrphanedPosterRelations(ctx)
	if err != nil {
		return err
	}
	for _, rel := range orphans {
		if err := deletePoster(ctx, rel); err != nil {
			return err
		}
	}
	return nil
}

func deletePoster(ctx context.Context, rel posterRelation) error {
	if err := media.DeleteMediaRelation(ctx, "film", rel.EntityID, rel.Path); err != nil {
		return fmt.Errorf("failed to delete poster relation %s: %w", rel.Path, err)
	}
	if err := media.DeleteMedia(ctx, rel.MediaID); err != nil {
		return fmt.Errorf("failed to delete poster media %d: %w", rel.MediaID, err)
	}
	return nil
}

func imageDimensions(data []byte) (width, height *int) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, nil
	}
	return &cfg.Width, &cfg.Height
}

func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
