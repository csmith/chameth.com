package contact

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	fromAddress   = flag.String("contact-from", "", "address to send e-mail from")
	toAddress     = flag.String("contact-to", "", "address to send e-mail to")
	emailSubject  = flag.String("contact-subject", "Contact form submission", "e-mail subject")
	smtpServer    = flag.String("contact-smtp-host", "", "SMTP server to connect to")
	smtpPort      = flag.Int("contact-smtp-port", 25, "port to use when connecting to the SMTP server")
	smtpUsername  = flag.String("contact-smtp-user", "", "username to supply to the SMTP server")
	smtpPassword  = flag.String("contact-smtp-pass", "", "password to supply to the SMTP server")
	signingSecret = flag.String("contact-signing-secret", "", "secret key used to sign form timestamps")
	oopspamApiKey = flag.String("oopspam-apikey", "", "OOPSpam API key (empty to disable spam checks on form submissions)")

	rateLimitMu  sync.Mutex
	rateLimitMap = make(map[string]time.Time)
	rateLimitTTL = 1 * time.Minute

	minFormAge = 10 * time.Second
)

func Process(ctx context.Context, req Request, method Method, remoteAddr, userAgent string) error {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	var failedChecks []cause

	for _, check := range checks {
		err := check(req, host)
		if err != nil {
			var rej *rejection
			if errors.As(err, &rej) {
				failedChecks = append(failedChecks, rej.cause)
			} else {
				slog.Error("Error checking contact form for spam", "request", req, "error", err)
			}
		}
	}

	if method == MethodForm && len(failedChecks) == 0 {
		err := checkOOPSpam(req, host)
		if err != nil {
			var rej *rejection
			if errors.As(err, &rej) {
				failedChecks = append(failedChecks, rej.cause)
			} else {
				slog.Error("Error checking contact form for spam", "request", req, "error", err)
			}
		}
	}

	recordMetric(ctx, string(method), userAgent, remoteAddr, failedChecks, req)

	if len(failedChecks) > 0 {
		time.Sleep(5 * time.Second)
		return ErrRejected
	}

	content := messageBody(req, method, remoteAddr)
	if err := sendContact(req, content); err != nil {
		slog.Error("Error sending contact form", "error", err, "request", req)
		return fmt.Errorf("failed to send: %w", err)
	}

	return nil
}

func sendContact(req Request, content string) error {
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

func messageBody(c Request, method Method, remoteAddr string) string {
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
	body.WriteString(string(method))
	body.WriteString("\n")

	body.WriteString("\nMESSAGE:\n\n")
	body.WriteString(c.Message)
	return body.String()
}

func SignedTimestamp() string {
	ts := time.Now().Unix()
	mac := hmac.New(sha256.New, []byte(*signingSecret))
	mac.Write([]byte(strconv.FormatInt(ts, 10)))
	return fmt.Sprintf("%d.%s", ts, hex.EncodeToString(mac.Sum(nil)))
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

func isRateAllowed(ip string) bool {
	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()

	if lastSubmission, exists := rateLimitMap[ip]; exists && time.Since(lastSubmission) < rateLimitTTL {
		return false
	}

	rateLimitMap[ip] = time.Now()
	return true
}
