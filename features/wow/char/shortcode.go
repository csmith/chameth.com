package char

import (
	"context"
	"fmt"
	"time"

	"chameth.com/chameth.com/external/ogrestream"
	"chameth.com/chameth.com/features/shortcodes"
	"tailscale.com/tsnet"
)

// refreshFrequency is how often character data is refreshed from the Ogre
// Stream API.
const refreshFrequency = 4 * time.Hour

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	shortcodes.RegisterData(mgr, "wowchar", 1, &dataShortcode{ts: ts})
}

// dataShortcode fetches a character's current (or historical) data from
// the Ogre Stream API, via the shortcodes data cache.
type dataShortcode struct {
	ts *tsnet.Server
}

// parseArgs interprets the shortcode's arguments: realm and character,
// plus an optional YYYY-MM-DD date to read the character's data as of.
func parseArgs(args []string) (realm, name string, at time.Time, err error) {
	if len(args) < 2 || len(args) > 3 {
		return "", "", time.Time{}, fmt.Errorf("wowchar requires 2 or 3 arguments (realm character [date])")
	}

	realm, name = args[0], args[1]
	if len(args) == 3 {
		at, err = time.Parse("2006-01-02", args[2])
		if err != nil {
			return "", "", time.Time{}, fmt.Errorf("invalid date: %s (expected YYYY-MM-DD)", args[2])
		}
	}

	return realm, name, at, nil
}

// Retrieve fetches the character's data as of the requested moment,
// rehosting the portrait alongside it.
func (s *dataShortcode) Retrieve(ctx context.Context, args []string) (shortcodes.Retrieved[Data], error) {
	realm, name, at, err := parseArgs(args)
	if err != nil {
		return shortcodes.Retrieved[Data]{}, err
	}

	client := ogrestream.NewClient(s.ts.HTTPClient())

	atParam := ""
	if !at.IsZero() {
		atParam = at.Format("2006-01-02")
	}

	c, err := client.Character(ctx, realm, name, atParam)
	if err != nil {
		return shortcodes.Retrieved[Data]{}, fmt.Errorf("failed to fetch character: %w", err)
	}
	if c.Profile == nil {
		return shortcodes.Retrieved[Data]{}, fmt.Errorf("character %s-%s has no profile data", name, realm)
	}

	imagePath, err := ensurePortrait(ctx, client, c, at)
	if err != nil {
		return shortcodes.Retrieved[Data]{}, fmt.Errorf("failed to rehost portrait: %w", err)
	}

	return shortcodes.Retrieved[Data]{Data: buildData(c, imagePath)}, nil
}

// RefreshPolicy refreshes every refreshFrequency. A dated request stops
// refreshing once its moment has passed: the service's answer for a past
// date can no longer change.
func (s *dataShortcode) RefreshPolicy(args []string) shortcodes.RefreshPolicy {
	policy := shortcodes.RefreshPolicy{Frequency: refreshFrequency}
	if _, _, at, err := parseArgs(args); err == nil && !at.IsZero() {
		policy.Cutoff = at
	}
	return policy
}

// Render builds the character card from the cached data.
func (s *dataShortcode) Render(_ []string, data Data, _ *shortcodes.Context) (string, error) {
	return renderTemplate(data)
}
