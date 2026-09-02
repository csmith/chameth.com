package shortcodes

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"
)

// RegisterGoroutine launches the background refresher for cached shortcode
// data. Wired automatically by cmd/generate.
func RegisterGoroutine(mgr *Manager, ctx context.Context) func() {
	return func() {
		refreshDueData(mgr, ctx)

		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				refreshDueData(mgr, ctx)
			}
		}
	}
}

func refreshDueData(mgr *Manager, ctx context.Context) {
	entries, err := dueShortcodeData(ctx)
	if err != nil {
		slog.Error("Failed to list due shortcode data", "error", err)
		return
	}

	for _, entry := range entries {
		reg, ok := mgr.data[entry.Shortcode]
		if !ok || reg.version != entry.Version {
			// Leftover row from a removed shortcode or an old version.
			// Leave it alone: versioning prevents it ever being rendered.
			continue
		}

		var args []string
		if err := json.Unmarshal(entry.Args, &args); err != nil {
			slog.Error("Failed to decode shortcode data args", "shortcode", entry.Shortcode, "error", err)
			continue
		}

		_, err = mgr.fetchData(ctx, entry.Shortcode, reg, args, entry.ArgsHash, true)
		if errors.Is(err, errFetchInProgress) {
			continue
		}
		if err != nil {
			// fetchData negative-caches the failure: existing data is
			// untouched and the row is rescheduled for a background retry.
			slog.Error("Failed to refresh shortcode data", "shortcode", entry.Shortcode, "error", err)
		}
	}
}
