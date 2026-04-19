package metrics

import "context"

type ContactSubmission struct {
	Method       string
	UserAgent    string
	RemoteAddr   string
	FailedChecks []string
	Page         string
	SenderName   string
	SenderEmail  string
	Message      string
}

func RecordContactSubmission(ctx context.Context, sub ContactSubmission) {
	recordContactMetric(ctx, sub)
}
