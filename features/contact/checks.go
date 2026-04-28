package contact

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"chameth.com/chameth.com/external/spamhaus"
)

type check = func(req request, remoteAddr string) error

var checks = []check{checkHoneypot, checkTimestamp, checkRateLimit, checkSensible, checkCyrillic, checkUnsubscribeLink, checkSpamhaus}

func checkHoneypot(req request, _ string) error {
	if req.Honeypot != "" {
		slog.Info("Honeypot field filled in contact form submission", "subject", req.Honeypot)
		return &rejection{cause: causeHoneypot}
	}
	return nil
}

func checkTimestamp(req request, _ string) error {
	if req.Timestamp == "" {
		slog.Info("Missing timestamp in contact form submission")
		return &rejection{cause: causeTimestampInvalid}
	}

	tsStr, sig, ok := strings.Cut(req.Timestamp, ".")
	if !ok {
		slog.Info("Malformed timestamp in contact form submission", "timestamp", req.Timestamp)
		return &rejection{cause: causeTimestampInvalid}
	}

	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		slog.Info("Unparseable timestamp in contact form submission", "timestamp", req.Timestamp)
		return &rejection{cause: causeTimestampInvalid}
	}

	mac := hmac.New(sha256.New, []byte(*signingSecret))
	mac.Write([]byte(tsStr))
	if !hmac.Equal([]byte(sig), []byte(hex.EncodeToString(mac.Sum(nil)))) {
		slog.Info("Invalid timestamp signature in contact form submission", "timestamp", req.Timestamp)
		return &rejection{cause: causeTimestampInvalid}
	}

	if elapsed := time.Since(time.Unix(ts, 0)); elapsed < minFormAge {
		slog.Info("Contact form submitted too quickly", "elapsed", elapsed)
		return &rejection{cause: causeTimestampTooSoon}
	}
	return nil
}

func checkRateLimit(req request, remoteAddr string) error {
	if !isRateAllowed(remoteAddr) {
		slog.Info("Rate limit exceeded for contact form", "remoteAddr", remoteAddr, "request", req)
		return &rejection{cause: causeRateLimit}
	}
	return nil
}

func checkSensible(req request, _ string) error {
	trimmed := strings.TrimSpace(req.Message)
	if trimmed != "" && len(strings.Fields(trimmed)) >= 2 {
		return nil
	}
	slog.Info("Blocking nonsensical contact form message", "request", req)
	return &rejection{cause: causeSensible}
}

func checkCyrillic(req request, _ string) error {
	for _, r := range req.Message {
		if r >= '\u0400' && r <= '\u04FF' {
			slog.Info("Blocking Cyrillic contact form message", "request", req)
			return &rejection{cause: causeCyrillic}
		}
	}
	return nil
}

func checkUnsubscribeLink(req request, _ string) error {
	if strings.Contains(req.Message, "unsubscribe.php?d=chameth.com") {
		slog.Info("Blocking unsubscribe link in contact form message", "request", req)
		return &rejection{cause: causeUnsubscribeLink}
	}
	return nil
}

func checkSpamhaus(req request, remoteAddr string) error {
	result, err := spamhaus.Check(remoteAddr)
	if err != nil {
		slog.Error("Error checking Spamhaus", "error", err, "remoteAddr", remoteAddr)
		return nil
	}

	if result.ExploitsBlockList {
		slog.Info("Blocking contact form from XBL listed address", "remoteAddr", remoteAddr, "request", req)
		return &rejection{cause: causeSpamhaus}
	}
	return nil
}
