package metrics

import "context"

type ContactSubmission struct {
	// WorkflowID identifies the submission, so a retried insert doesn't add
	// a second row.
	WorkflowID   string
	Method       string
	UserAgent    string
	RemoteAddr   string
	FailedChecks []string
	Page         string
	SenderName   string
	SenderEmail  string
	Message      string
	LLMReason    *string
}

func RecordContactSubmission(ctx context.Context, sub ContactSubmission) error {
	return recordContactMetric(ctx, sub)
}
