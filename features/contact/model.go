package contact

import "time"

type method string

const (
	methodJSON method = "JSON"
	methodForm method = "Form"
)

type request struct {
	Page        string `json:"page"`
	SenderName  string `json:"name"`
	SenderEmail string `json:"email"`
	Message     string `json:"message"`
	Timestamp   string `json:"ts"`
	Honeypot    string `json:"subject"`
}

// submission is the contact workflow's input: the request plus what the
// handler knew about where and when it came from.
type submission struct {
	Request    request   `json:"request"`
	Method     method    `json:"method"`
	RemoteAddr string    `json:"remoteAddr"`
	UserAgent  string    `json:"userAgent"`
	ReceivedAt time.Time `json:"receivedAt"`
}

// spamVerdict is the check-blog-spam workflow's output.
type spamVerdict struct {
	Send   bool   `json:"send"`
	Reason string `json:"reason"`
}

type cause string

const (
	causeHoneypot         cause = "honeypot_populated"
	causeTimestampInvalid cause = "timestamp_invalid"
	causeTimestampTooSoon cause = "timestamp_too_soon"
	causeRateLimit        cause = "rate_limit_exceeded"
	causeSensible         cause = "nonsense_message"
	causeCyrillic         cause = "contained_cyrillic"
	causeSpamhaus         cause = "listed_in_xbl"
	causeUnsubscribeLink  cause = "unsubscribe_link"
	causeLLM              cause = "llm_spam"
	causeDuplicate        cause = "duplicate_message"
)

type rejection struct {
	cause cause
}

func (e *rejection) Error() string {
	return "submission rejected: " + string(e.cause)
}
