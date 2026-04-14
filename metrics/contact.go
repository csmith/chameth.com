package metrics

// ContactRejectionCause represents a reason a contact form submission was rejected.
type ContactRejectionCause string

const (
	CauseHoneypot         ContactRejectionCause = "honeypot_populated"
	CauseTimestampInvalid ContactRejectionCause = "timestamp_invalid"
	CauseTimestampTooSoon ContactRejectionCause = "timestamp_too_soon"
	CauseRateLimit        ContactRejectionCause = "rate_limit_exceeded"
	CauseSensible         ContactRejectionCause = "nonsense_message"
	CauseCyrillic         ContactRejectionCause = "contained_cyrillic"
	CauseSpamhaus         ContactRejectionCause = "listed_in_xbl"
	CauseOOPSpam          ContactRejectionCause = "flagged_by_oopspam"
)

var contactRejectionCauses = []ContactRejectionCause{
	CauseHoneypot,
	CauseTimestampInvalid,
	CauseTimestampTooSoon,
	CauseRateLimit,
	CauseSensible,
	CauseCyrillic,
	CauseSpamhaus,
	CauseOOPSpam,
}

func contactCauseLabelNames() []string {
	names := make([]string, len(contactRejectionCauses))
	for i, c := range contactRejectionCauses {
		names[i] = string(c)
	}
	return names
}

func RecordContactSubmission(method string, causes []ContactRejectionCause) {
	causeSet := make(map[ContactRejectionCause]bool, len(causes))
	for _, c := range causes {
		causeSet[c] = true
	}
	labels := make([]string, len(contactRejectionCauses)+1)
	labels[0] = method
	for i, c := range contactRejectionCauses {
		if causeSet[c] {
			labels[i+1] = "true"
		} else {
			labels[i+1] = "false"
		}
	}
	contactSubmissionsTotal.WithLabelValues(labels...).Inc()
}
