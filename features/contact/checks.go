package contact

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"chameth.com/chameth.com/external/oopspam"
	"chameth.com/chameth.com/external/spamhaus"
)

type check = func(req Request, remoteAddr string) error

var checks = map[Method][]check{
	MethodJSON: {checkRateLimit, checkSensibleMessage, checkCyrillic, checkSpamhaus},
	MethodForm: {checkRateLimit, checkSensibleMessage, checkCyrillic, checkSpamhaus, checkOOPSpam},
}

var (
	errRateLimitExceeded  = errors.New("rate limit exceeded")
	errNonsensicalMessage = errors.New("nonsensical message")
	errCyrillicMessage    = errors.New("cyrillic message")
	errBlockedBySpamhaus  = errors.New("blocked by spamhaus")
	errSpamDetected       = errors.New("spam detected")
)

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
