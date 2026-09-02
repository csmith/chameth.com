package shortcodes

import (
	"context"
	"time"

	"chameth.com/chameth.com/db"
)

func getShortcodeData(ctx context.Context, shortcode string, version int, argsHash string) (*shortcodeData, error) {
	entry, err := db.Get[shortcodeData](ctx, `
		SELECT id, shortcode, version, args_hash, args, data, next_refresh_at, retrieved_at
		FROM shortcode_data
		WHERE shortcode = $1 AND version = $2 AND args_hash = $3
	`, shortcode, version, argsHash)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func upsertShortcodeData(ctx context.Context, shortcode string, version int, argsHash string, argsJSON, dataJSON []byte, retrievedAt time.Time, nextRefreshAt *time.Time) error {
	_, err := db.Exec(ctx, `
		INSERT INTO shortcode_data (shortcode, version, args_hash, args, data, retrieved_at, next_refresh_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (shortcode, version, args_hash) DO UPDATE SET
			args = EXCLUDED.args,
			data = EXCLUDED.data,
			retrieved_at = EXCLUDED.retrieved_at,
			next_refresh_at = EXCLUDED.next_refresh_at
	`, shortcode, version, argsHash, string(argsJSON), string(dataJSON), retrievedAt, nextRefreshAt)
	return err
}

// upsertShortcodeDataFailure records a failed retrieval: on a new row it
// stores no data; on an existing row it leaves data and retrieved_at
// untouched and only pushes the retry time forward.
func upsertShortcodeDataFailure(ctx context.Context, shortcode string, version int, argsHash string, argsJSON []byte, failedAt time.Time, nextRefreshAt *time.Time) error {
	_, err := db.Exec(ctx, `
		INSERT INTO shortcode_data (shortcode, version, args_hash, args, data, retrieved_at, next_refresh_at)
		VALUES ($1, $2, $3, $4, NULL, $5, $6)
		ON CONFLICT (shortcode, version, args_hash) DO UPDATE SET
			next_refresh_at = EXCLUDED.next_refresh_at
	`, shortcode, version, argsHash, string(argsJSON), failedAt, nextRefreshAt)
	return err
}

func dueShortcodeData(ctx context.Context) ([]shortcodeData, error) {
	return db.Select[shortcodeData](ctx, `
		SELECT id, shortcode, version, args_hash, args, data, next_refresh_at, retrieved_at
		FROM shortcode_data
		WHERE next_refresh_at IS NOT NULL AND next_refresh_at <= now()
		ORDER BY retrieved_at
	`)
}
