package contact

import (
	"fmt"
	"time"
	"unicode/utf8"
)

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

const (
	// Maximum field lengths, in characters. The form renders the user-editable
	// ones as maxlength attributes so browsers enforce them before submitting.
	MaxNameLength    = 200
	MaxEmailLength   = 254
	MaxMessageLength = 10000
	maxPageLength    = 1024
	maxOtherLength   = 256
)

func (r request) validateLengths() error {
	fields := []struct {
		name  string
		value string
		max   int
	}{
		{"page", r.Page, maxPageLength},
		{"name", r.SenderName, MaxNameLength},
		{"email", r.SenderEmail, MaxEmailLength},
		{"message", r.Message, MaxMessageLength},
		{"ts", r.Timestamp, maxOtherLength},
		{"subject", r.Honeypot, maxOtherLength},
	}
	for _, f := range fields {
		if n := utf8.RuneCountInString(f.value); n > f.max {
			return fmt.Errorf("%s is %d characters, maximum is %d", f.name, n, f.max)
		}
	}
	return nil
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

type checkPairCount struct {
	RowCheck string `db:"row_check"`
	ColCheck string `db:"col_check"`
	Count    int    `db:"count"`
}
