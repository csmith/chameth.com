package contact

import (
	"context"
	"encoding/json"
	"log/slog"

	"chameth.com/chameth.com/db"
)

func recordMetric(ctx context.Context, method, userAgent, remoteAddr string, failedChecks []cause, req Request) {
	checksJSON, err := json.Marshal(failedChecks)
	if err != nil {
		slog.Error("Error marshalling contact checks", "error", err)
		return
	}

	_, err = db.NamedExec(ctx, `
		INSERT INTO contact_metrics (method, user_agent, remote_addr, checks, page, sender_name, sender_email, message)
		VALUES (:method, :user_agent, :remote_addr, :checks, :page, :sender_name, :sender_email, :message)
	`, map[string]any{
		"method":       method,
		"user_agent":   userAgent,
		"remote_addr":  remoteAddr,
		"checks":       checksJSON,
		"page":         req.Page,
		"sender_name":  req.SenderName,
		"sender_email": req.SenderEmail,
		"message":      req.Message,
	})
	if err != nil {
		slog.Error("Error recording contact metric", "error", err)
	}
}
