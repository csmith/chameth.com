package contact

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log/slog"
	"net/mail"
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
	rateLimitMu   sync.Mutex
	rateLimitMap  = make(map[string]time.Time)
	rateLimitTTL  = 1 * time.Minute

	minFormAge = 10 * time.Second
)

func sendContact(req request, content string) error {
	auth := smtp.PlainAuth("", *smtpUsername, *smtpPassword, *smtpServer)
	// ParseAddress rejects anything unsafe to embed in a header (CRLF, control chars)
	replyTo := "noreply@chameth.com"
	if addr, err := mail.ParseAddress(req.SenderEmail); err == nil {
		replyTo = addr.Address
	}
	body := fmt.Sprintf(
		"To: %s\r\nSubject: %s\r\nReply-to: %s\r\nFrom: Online contact form <%s>\r\n\r\n%s\r\n",
		*toAddress,
		*emailSubject,
		replyTo,
		*fromAddress,
		content,
	)
	slog.Info(
		"Sending e-mail message",
		"from",
		*fromAddress,
		"to",
		*toAddress,
		"subject",
		*emailSubject,
		"replyTo",
		req.SenderEmail,
	)
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", *smtpServer, *smtpPort),
		auth,
		*fromAddress,
		[]string{*toAddress},
		[]byte(body),
	)
	if err != nil {
		slog.Error("Unable to send e-mail", "error", err)
		return err
	}
	return nil
}

func messageBody(c request, mthd method, remoteAddr string) string {
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
	body.WriteString(string(mthd))
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
