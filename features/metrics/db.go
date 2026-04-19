package metrics

import (
	"context"
	"encoding/json"
	"log/slog"

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
