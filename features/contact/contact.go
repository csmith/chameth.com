package contact

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"chameth.com/chameth.com/external/oopspam"
	"chameth.com/chameth.com/external/spamhaus"
)

var (
	fromAddress   = flag.String("contact-from", "", "address to send e-mail from")
	toAddress     = flag.String("contact-to", "", "address to send e-mail to")
	emailSubject  = flag.String("contact-subject", "Contact form submission", "e-mail subject")
	smtpServer    = flag.String("contact-smtp-host", "", "SMTP server to connect to")
	smtpPort      = flag.Int("contact-smtp-port", 25, "port to use when connecting to the SMTP server")
	smtpUsername  = flag.String("contact-smtp-user", "", "username to supply to the SMTP server")
	smtpPassword  = flag.String("contact-smtp-pass", "", "password to supply to the SMTP server")
	oopspamApiKey = flag.String("oopspam-apikey", "", "OOPSpam API key (empty to disable spam checks on form submissions)")

	rateLimitMu  sync.Mutex
	rateLimitMap = make(map[string]time.Time)
	rateLimitTTL = 1 * time.Minute
)

// Rejection is returned when a submission is rejected by a spam or abuse check.
// Use errors.As to check for this type to distinguish rejections from internal errors.
type Rejection struct {
	Err error
}

func (e *Rejection) Error() string { return e.Err.Error() }
func (e *Rejection) Unwrap() error { return e.Err }

// Sentinel errors for specific rejection reasons.
var (
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrNonsensicalMessage = errors.New("nonsensical message")
	ErrCyrillicMessage    = errors.New("cyrillic message")
	ErrBlockedBySpamhaus  = errors.New("blocked by spamhaus")
	ErrSpamDetected       = errors.New("spam detected")
)

// Method indicates how the contact form submission was received.
type Method int

const (
	MethodJSON Method = iota
	MethodForm
)

func (m Method) String() string {
	switch m {
	case MethodJSON:
		return "json"
	case MethodForm:
		return "form"
	default:
		return "unknown"
	}
}

// ContactRequest holds the data from a contact form submission.
type ContactRequest struct {
	Page        string
	SenderName  string
	SenderEmail string
	Message     string
}

// Process validates the submission, runs spam checks, and sends the email.
// It returns a *Rejection for spam/abuse rejections, or a regular error for
// internal failures (e.g. email send failure).
func Process(req ContactRequest, method Method, remoteAddr string) error {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	if !checkRateLimit(host) {
		slog.Info("Rate limit exceeded for contact form", "remoteAddr", host, "request", req)
		time.Sleep(5 * time.Second)
		return &Rejection{Err: ErrRateLimitExceeded}
	}

	if !isSensibleMessage(req.Message) {
		slog.Info("Blocking nonsensical contact form message", "remoteAddr", host, "request", req)
		time.Sleep(5 * time.Second)
		return &Rejection{Err: ErrNonsensicalMessage}
	}

	if containsCyrillic(req.Message) {
		slog.Info("Blocking Cyrillic contact form message", "remoteAddr", host, "request", req)
		time.Sleep(5 * time.Second)
		return &Rejection{Err: ErrCyrillicMessage}
	}

	spam, err := spamhaus.Check(host)
	if err != nil {
		slog.Error("Error checking Spamhaus", "error", err, "remoteAddr", host)
	}

	if spam.ExploitsBlockList {
		slog.Info("Blocking contact form from XBL listed address", "remoteAddr", host, "request", req)
		time.Sleep(5 * time.Second)
		return &Rejection{Err: ErrBlockedBySpamhaus}
	}

	if method == MethodForm && *oopspamApiKey != "" {
		result, err := oopspam.IsSpam(*oopspamApiKey, req.Message, host, req.SenderEmail)
		if err != nil {
			slog.Error("OOPSpam check failed", "error", err, "remoteAddr", host)
			return fmt.Errorf("oopspam check: %w", err)
		}

		if result.IsSpam {
			slog.Info("OOPSpam detected spam, blocking submission", "remoteAddr", host, "score", result.Score, "details", result.Details)
			return &Rejection{Err: ErrSpamDetected}
		}
	}

	content := messageBody(req, method, remoteAddr)
	if err := sendContact(req, content); err != nil {
		slog.Error("Error sending contact form", "error", err, "request", req)
		return fmt.Errorf("failed to send: %w", err)
	}

	return nil
}

func sendContact(req ContactRequest, content string) error {
	auth := smtp.PlainAuth("", *smtpUsername, *smtpPassword, *smtpServer)
	replyTo := req.SenderEmail
	if replyTo == "" {
		replyTo = "noreply@chameth.com"
	}
	body := fmt.Sprintf("To: %s\r\nSubject: %s\r\nReply-to: %s\r\nFrom: Online contact form <%s>\r\n\r\n%s\r\n", *toAddress, *emailSubject, replyTo, *fromAddress, content)
	slog.Info("Sending e-mail message", "from", *fromAddress, "to", *toAddress, "subject", *emailSubject, "replyTo", req.SenderEmail)
	err := smtp.SendMail(fmt.Sprintf("%s:%d", *smtpServer, *smtpPort), auth, *fromAddress, []string{*toAddress}, []byte(body))
	if err != nil {
		slog.Error("Unable to send e-mail", "error", err)
		return err
	}
	return nil
}

func messageBody(c ContactRequest, method Method, remoteAddr string) string {
	body := strings.Builder{}
	body.WriteString("SENDER: ")
	body.WriteString(c.SenderName)
	body.WriteString(" <")
	body.WriteString(c.SenderEmail)
	body.WriteString(">\n\n")
	body.WriteString("PAGE: ")
	body.WriteString(c.Page)
	body.WriteString("\n\n")
	body.WriteString("REMOTEIP: ")
	body.WriteString(remoteAddr)
	body.WriteString("\n")
	body.WriteString("METHOD: ")
	body.WriteString(method.String())
	body.WriteString("\n")

	body.WriteString("\nMESSAGE:\n\n")
	body.WriteString(c.Message)
	return body.String()
}

func isSensibleMessage(message string) bool {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return false
	}
	return len(strings.Fields(trimmed)) >= 2
}

func containsCyrillic(s string) bool {
	for _, r := range s {
		if r >= '\u0400' && r <= '\u04FF' {
			return true
		}
	}
	return false
}

func init() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			cleanupRateLimitMap()
		}
	}()
}

func cleanupRateLimitMap() {
	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()

	now := time.Now()
	for ip, lastSeen := range rateLimitMap {
		if now.Sub(lastSeen) > 5*time.Minute {
			delete(rateLimitMap, ip)
		}
	}
}

func checkRateLimit(ip string) bool {
	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()

	lastSubmission, exists := rateLimitMap[ip]
	if exists && time.Since(lastSubmission) < rateLimitTTL {
		return false
	}

	rateLimitMap[ip] = time.Now()
	return true
}
