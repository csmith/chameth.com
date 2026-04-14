package contact

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"chameth.com/chameth.com/external/oopspam"
	"chameth.com/chameth.com/external/spamhaus"
)

type check = func(req Request, remoteAddr string) error

var commonChecks = []check{checkHoneypot, checkTimestamp, checkRateLimit, checkSensibleMessage, checkCyrillic, checkSpamhaus}

var checks = map[Method][]check{
	MethodJSON: commonChecks,
	MethodForm: append(slices.Clone(commonChecks), checkOOPSpam),
}

var (
	errRateLimitExceeded  = errors.New("rate limit exceeded")
	errNonsensicalMessage = errors.New("nonsensical message")
	errCyrillicMessage    = errors.New("cyrillic message")
	errBlockedBySpamhaus  = errors.New("blocked by spamhaus")
	errSpamDetected       = errors.New("spam detected")
	errInvalidTimestamp   = errors.New("invalid timestamp")
	errTimestampTooRecent = errors.New("timestamp too recent")
	errHoneypotFilled     = errors.New("honeypot field filled")
)

func checkHoneypot(req Request, _ string) error {
	if req.Honeypot != "" {
		slog.Info("Honeypot field filled in contact form submission", "subject", req.Honeypot)
		return &Rejection{Err: errHoneypotFilled}
	}
	return nil
}

func checkTimestamp(req Request, _ string) error {
	if req.Timestamp == "" {
		slog.Info("Missing timestamp in contact form submission")
		return &Rejection{Err: errInvalidTimestamp}
	}

	tsStr, sig, ok := strings.Cut(req.Timestamp, ".")
	if !ok {
		slog.Info("Malformed timestamp in contact form submission", "timestamp", req.Timestamp)
		return &Rejection{Err: errInvalidTimestamp}
	}

	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		slog.Info("Unparseable timestamp in contact form submission", "timestamp", req.Timestamp)
		return &Rejection{Err: errInvalidTimestamp}
	}

	mac := hmac.New(sha256.New, []byte(*signingSecret))
	mac.Write([]byte(tsStr))
	if !hmac.Equal([]byte(sig), []byte(hex.EncodeToString(mac.Sum(nil)))) {
		slog.Info("Invalid timestamp signature in contact form submission", "timestamp", req.Timestamp)
		return &Rejection{Err: errInvalidTimestamp}
	}

	if elapsed := time.Since(time.Unix(ts, 0)); elapsed < minFormAge {
		slog.Info("Contact form submitted too quickly", "elapsed", elapsed)
		return &Rejection{Err: errTimestampTooRecent}
	}
	return nil
}

func checkRateLimit(req Request, remoteAddr string) error {
	if !isRateAllowed(remoteAddr) {
		slog.Info("Rate limit exceeded for contact form", "remoteAddr", remoteAddr, "request", req)
		return &Rejection{Err: errRateLimitExceeded}
	}
	return nil
}

func checkSensibleMessage(req Request, _ string) error {
	trimmed := strings.TrimSpace(req.Message)
	if trimmed != "" && len(strings.Fields(trimmed)) >= 2 {
		return nil
	}
	slog.Info("Blocking nonsensical contact form message", "request", req)
	return &Rejection{Err: errNonsensicalMessage}
}

func checkCyrillic(req Request, _ string) error {
	for _, r := range req.Message {
		if r >= '\u0400' && r <= '\u04FF' {
			slog.Info("Blocking Cyrillic contact form message", "request", req)
			return &Rejection{Err: errCyrillicMessage}
		}
	}
	return nil
}

func checkSpamhaus(req Request, remoteAddr string) error {
	result, err := spamhaus.Check(remoteAddr)
	if err != nil {
		slog.Error("Error checking Spamhaus", "error", err, "remoteAddr", remoteAddr)
		return nil
	}

	if result.ExploitsBlockList {
		slog.Info("Blocking contact form from XBL listed address", "remoteAddr", remoteAddr, "request", req)
		return &Rejection{Err: errBlockedBySpamhaus}
	}
	return nil
}

func checkOOPSpam(req Request, remoteAddr string) error {
	if *oopspamApiKey == "" {
		return nil
	}

	result, err := oopspam.IsSpam(*oopspamApiKey, req.Message, remoteAddr, req.SenderEmail)
	if err != nil {
		return fmt.Errorf("oopspam check: %w", err)
	}

	if result.IsSpam {
		slog.Info("OOPSpam detected spam, blocking submission", "remoteAddr", remoteAddr, "score", result.Score, "details", result.Details)
		return &Rejection{Err: errSpamDetected}
	}
	return nil
}
