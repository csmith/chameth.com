package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

var (
	fromAddress  = flag.String("contact-from", "", "address to send e-mail from")
	toAddress    = flag.String("contact-to", "", "address to send e-mail to")
	subject      = flag.String("contact-subject", "Contact form submission", "e-mail subject")
	smtpServer   = flag.String("contact-smtp-host", "", "SMTP server to connect to")
	smtpPort     = flag.Int("contact-smtp-port", 25, "port to use when connecting to the SMTP server")
	smtpUsername = flag.String("contact-smtp-user", "", "username to supply to the SMTP server")
	smtpPassword = flag.String("contact-smtp-pass", "", "password to supply to the SMTP server")
)

type contactRequest struct {
	Page        string `json:"page"`
	SenderName  string `json:"name"`
	SenderEmail string `json:"email"`
	Message     string `json:"message"`
}

var (
	rateLimitMu  sync.Mutex
	rateLimitMap = make(map[string]time.Time)
	rateLimitTTL = 1 * time.Minute
)

func init() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			cleanupRateLimitMap()
		}
	}()
}

func ContactForm(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Error reading contact form body", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var cr contactRequest
	if err = json.Unmarshal(body, &cr); err != nil {
		slog.Error("Error parsing contact form payload", "error", err, "payload", string(body))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	if !checkRateLimit(host) {
		slog.Info("Rate limit exceeded for contact form", "remoteAddr", host, "request", cr)
		time.Sleep(5 * time.Second)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if isRandomCharacterSpam(cr.SenderName, cr.Message) {
		slog.Info("Blocking random character spam", "remoteAddr", host, "request", cr)
		time.Sleep(5 * time.Second)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	spam, err := checkSpamhaus(host)
	if err != nil {
		slog.Error("Error checking Spamhaus", "error", err, "remoteAddr", host)
	}

	if spam.ExploitsBlockList {
		slog.Info("Blocking contact form from XBL listed address", "remoteAddr", host, "request", cr)
		time.Sleep(5 * time.Second)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if err := sendContact(cr, messageBody(cr, r, spam)); err != nil {
		slog.Error("Error sending contact form", "error", err, "request", cr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func sendContact(req contactRequest, content string) error {
	auth := smtp.PlainAuth("", *smtpUsername, *smtpPassword, *smtpServer)
	body := fmt.Sprintf("To: %s\r\nSubject: %s\r\nReply-to: %s\r\nFrom: Online contact form <%s>\r\n\r\n%s\r\n", *toAddress, *subject, req.SenderEmail, *fromAddress, content)
	slog.Info("Sending e-mail message", "from", *fromAddress, "to", *toAddress, "subject", *subject, "replyTo", req.SenderEmail)
	err := smtp.SendMail(fmt.Sprintf("%s:%d", *smtpServer, *smtpPort), auth, *fromAddress, []string{*toAddress}, []byte(body))
	if err != nil {
		slog.Error("Unable to send e-mail", "error", err)
		return err
	}
	return nil
}

func messageBody(c contactRequest, req *http.Request, result spamhausResult) string {
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
	body.WriteString(req.RemoteAddr)
	body.WriteString("\n")
	body.WriteString("SPAMHAUS: ")
	body.WriteString(result.Summary())
	body.WriteString("; see ")
	body.WriteString(result.CheckURL)
	body.WriteString("\n")

	body.WriteString("\nMESSAGE:\n\n")
	body.WriteString(c.Message)
	return body.String()
}

type spamhausResult struct {
	Success                 bool
	SpamhausBlockList       bool
	CombinedSpamSources     bool
	ExploitsBlockList       bool
	DontRouteOrPeer         bool
	PolicyBlockListISP      bool
	PolicyBlockListSpamhaus bool
	CheckURL                string
}

func (r spamhausResult) Summary() string {
	if !r.Success {
		return "check failed"
	}

	var lists []string
	if r.SpamhausBlockList {
		lists = append(lists, "SBL")
	}
	if r.CombinedSpamSources {
		lists = append(lists, "CSS")
	}
	if r.ExploitsBlockList {
		lists = append(lists, "XBL")
	}
	if r.DontRouteOrPeer {
		lists = append(lists, "DROP")
	}
	if r.PolicyBlockListISP {
		lists = append(lists, "PBL")
	}
	if r.PolicyBlockListSpamhaus {
		lists = append(lists, "PBL")
	}

	if len(lists) == 0 {
		return "not listed"
	}

	return "listed: " + strings.Join(lists, ", ")
}

func checkSpamhaus(host string) (spamhausResult, error) {
	result := spamhausResult{
		CheckURL: fmt.Sprintf("https://check.spamhaus.org/results/?query=%s", host),
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return result, fmt.Errorf("invalid IP address: %s", host)
	}

	var reversed string
	if ip.To4() != nil {
		reversed = reverseIPv4(ip)
	} else {
		reversed = reverseIPv6(ip)
	}

	if reversed == "" {
		return result, fmt.Errorf("failed to reverse IP: %s", host)
	}

	dnsQuery := reversed + ".zen.spamhaus.org"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resolver := &net.Resolver{}
	addrs, err := resolver.LookupHost(ctx, dnsQuery)

	if err != nil {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
			result.Success = true
			return result, nil
		}

		return result, fmt.Errorf("spamhaus DNS lookup failed: %w", err)
	}

	for _, addr := range addrs {
		switch addr {
		case "127.0.0.2":
			result.SpamhausBlockList = true
		case "127.0.0.3":
			result.CombinedSpamSources = true
		case "127.0.0.4":
			result.ExploitsBlockList = true
		case "127.0.0.9":
			result.DontRouteOrPeer = true
		case "127.0.0.10":
			result.PolicyBlockListISP = true
		case "127.0.0.11":
			result.PolicyBlockListSpamhaus = true
		}
	}

	result.Success = true
	return result, nil
}

// reverseIPv4 reverses the octets of an IPv4 address for DNSBL lookup
// Example: "204.12.215.98" -> "98.215.12.204"
func reverseIPv4(ip net.IP) string {
	ip = ip.To4()
	if ip == nil {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d.%d", ip[3], ip[2], ip[1], ip[0])
}

// reverseIPv6 reverses an IPv6 address for DNSBL lookup
// Example: "2001:db8::1" -> "1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2"
func reverseIPv6(ip net.IP) string {
	ip = ip.To16()
	if ip == nil {
		return ""
	}

	var parts []string
	// Convert each byte to two hex digits, then split them
	for i := len(ip) - 1; i >= 0; i-- {
		parts = append(parts, fmt.Sprintf("%x", ip[i]&0x0f))
		parts = append(parts, fmt.Sprintf("%x", ip[i]>>4))
	}

	return strings.Join(parts, ".")
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

func isRandomCharacterSpam(name, message string) bool {
	return !strings.Contains(name, " ") && !strings.Contains(message, " ")
}
