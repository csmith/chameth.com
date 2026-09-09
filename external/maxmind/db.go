package maxmind

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"chameth.com/chameth.com/db"
)

// replaceASNNetworks atomically replaces the contents of the asn_networks
// table. Everything happens in a single transaction, so if the download or
// any insert fails the previously loaded data is left untouched.
func lastRefresh(ctx context.Context) (time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var t time.Time
	row, err := db.Get[ASNRefresh](ctx, `SELECT refreshed_at FROM asn_refresh ORDER BY id DESC LIMIT 1`)
	t = row.RefreshedAt
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	return t, err
}

func replaceASNNetworks(ctx context.Context, networks []asnNetwork) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM asn_networks`); err != nil {
		return fmt.Errorf("failed to clear existing networks: %w", err)
	}

	const batchSize = 1000
	for start := 0; start < len(networks); start += batchSize {
		batch := networks[start:min(start+batchSize, len(networks))]

		var (
			query strings.Builder
			args  = make([]any, 0, len(batch)*3)
		)
		query.WriteString(`INSERT INTO asn_networks (network, asn, organization) VALUES `)
		for i, n := range batch {
			if i > 0 {
				query.WriteByte(',')
			}
			base := i * 3
			fmt.Fprintf(&query, "($%d::cidr, $%d, $%d)", base+1, base+2, base+3)
			args = append(args, n.network.String(), n.asn, n.organization)
		}
		query.WriteString(` ON CONFLICT (network) DO NOTHING`)

		if _, err = tx.ExecContext(ctx, query.String(), args...); err != nil {
			return fmt.Errorf("failed to insert networks: %w", err)
		}
	}

	if _, err = tx.ExecContext(ctx, `INSERT INTO asn_refresh (refreshed_at) VALUES (NOW())`); err != nil {
		return fmt.Errorf("failed to record ASN refresh: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit ASN update: %w", err)
	}
	return nil
}
