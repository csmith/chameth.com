package contact

import "errors"

type Method string

const (
	MethodJSON Method = "JSON"
	MethodForm Method = "Form"
)

type Request struct {
	Page        string `json:"page"`
	SenderName  string `json:"name"`
	SenderEmail string `json:"email"`
	Message     string `json:"message"`
	Timestamp   string `json:"ts"`
	Honeypot    string `json:"subject"`
}

var ErrRejected = errors.New("submission rejected")

type cause string

const (
	causeHoneypot         cause = "honeypot_populated"
	causeTimestampInvalid cause = "timestamp_invalid"
	causeTimestampTooSoon cause = "timestamp_too_soon"
	causeRateLimit        cause = "rate_limit_exceeded"
	causeSensible         cause = "nonsense_message"
	causeCyrillic         cause = "contained_cyrillic"
	causeSpamhaus         cause = "listed_in_xbl"
	causeOOPSpam          cause = "flagged_by_oopspam"
)

type rejection struct {
	cause cause
}

func (e *rejection) Error() string {
	return "submission rejected: " + string(e.cause)
}
