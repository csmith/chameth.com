package metrics

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"chameth.com/chameth.com/db"
)

func recordContactMetric(ctx context.Context, sub ContactSubmission) {
	checksJSON, err := json.Marshal(sub.FailedChecks)
	if err != nil {
		slog.Error("Error marshalling contact checks", "error", err)
		return
	}

	_, err = db.NamedExec(ctx, `
		INSERT INTO contact_metrics (method, user_agent, remote_addr, checks, page, sender_name, sender_email, message)
		VALUES (:method, :user_agent, :remote_addr, :checks, :page, :sender_name, :sender_email, :message)
	`, map[string]any{
		"method":       sub.Method,
		"user_agent":   sub.UserAgent,
		"remote_addr":  sub.RemoteAddr,
		"checks":       checksJSON,
		"page":         sub.Page,
		"sender_name":  sub.SenderName,
		"sender_email": sub.SenderEmail,
		"message":      sub.Message,
	})
	if err != nil {
		slog.Error("Error recording contact metric", "error", err)
	}
}

func insertRequestLog(ctx context.Context, log requestLog) {
	// The raw address is only ever a query parameter for the ASN containment
	// lookup; it is never written to a column. An invalid or unmapped
	// address yields a row with a NULL ASN, so exactly one row is always
	// inserted. Overlapping networks are resolved deterministically in
	// favour of the most specific (longest prefix) match.
	var addr any
	if log.ip.IsValid() {
		addr = log.ip.String()
	}
	_, err := db.Exec(ctx, `
		INSERT INTO request_logs (recorded_at, url, user_agent, ip_hash, asn, as_org, duration_us, response_bytes, status)
		SELECT $8, $1, $2, $3, n.asn, n.organization, $4, $5, $6
		FROM (SELECT $7::inet AS addr) q
		LEFT JOIN LATERAL (
			SELECT asn, organization
			FROM asn_networks
			WHERE network >>= q.addr
			ORDER BY masklen(network) DESC
			LIMIT 1
		) n ON true
	`, log.url, log.userAgent, log.ipHash, log.duration.Microseconds(), log.size, log.status, addr, log.start)
	if err != nil {
		slog.Error("Error inserting request log", "error", err)
	}
}

func recordRequestMetric(path, requestID string, duration time.Duration, queries int32) {
	if len(path) > 256 {
		path = path[:256]
	}

	_, err := db.NamedExec(context.Background(), `
		INSERT INTO request_metrics (path, request_id, duration_us, query_count)
		VALUES (:path, :request_id, :duration_us, :query_count)
	`, map[string]any{
		"path":        path,
		"request_id":  requestID,
		"duration_us": duration.Microseconds(),
		"query_count": queries,
	})
	if err != nil {
		slog.Error("Error recording request metric", "error", err)
	}
}
