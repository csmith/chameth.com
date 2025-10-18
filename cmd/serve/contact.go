package main

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

type ContactRequest struct {
	Page        string `json:"page"`
	SenderName  string `json:"name"`
	SenderEmail string `json:"email"`
	Message     string `json:"message"`
}

func handleContactForm(w http.ResponseWriter, r *http.Request) {
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

	var cr ContactRequest
	if err = json.Unmarshal(body, &cr); err != nil {
		slog.Error("Error parsing contact form payload", "error", err, "payload", string(body))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := sendContact(cr, messageBody(cr, r)); err != nil {
		slog.Error("Error sending contact form", "error", err, "request", cr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func sendContact(req ContactRequest, content string) error {
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

func messageBody(c ContactRequest, req *http.Request) string {
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
	body.WriteString(checkSpamhaus(req.RemoteAddr))
	body.WriteString("\n\n")
	body.WriteString("MESSAGE:\n\n")
	body.WriteString(c.Message)
	return body.String()
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

// checkSpamhaus queries the Spamhaus zen.spamhaus.org blocklist for the given IP
// Returns:
// - "Not listed" if the IP is not in the blocklist
// - "Listed - [reason]" if the IP is in the blocklist with the TXT record explanation
// - "Check failed" if the DNS lookup encounters an error
func checkSpamhaus(ipAddr string) string {
	host, _, err := net.SplitHostPort(ipAddr)
	if err != nil {
		// If SplitHostPort fails, assume it's just an IP without port
		host = ipAddr
	}

	ip := net.ParseIP(host)
	if ip == nil {
		slog.Warn("Invalid IP address for Spamhaus check", "ip", ipAddr)
		return "Check failed"
	}

	var reversed string
	if ip.To4() != nil {
		reversed = reverseIPv4(ip)
	} else {
		reversed = reverseIPv6(ip)
	}

	if reversed == "" {
		slog.Warn("Failed to reverse IP for Spamhaus check", "ip", ipAddr)
		return "Check failed"
	}

	dnsQuery := reversed + ".zen.spamhaus.org"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resolver := &net.Resolver{}
	txtRecords, err := resolver.LookupTXT(ctx, dnsQuery)

	if err != nil {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
			return "Not listed"
		}

		slog.Warn("Spamhaus DNS lookup failed", "query", dnsQuery, "error", err)
		return "Check failed"
	}

	if len(txtRecords) > 0 {
		return "Listed - " + strings.Join(txtRecords, "; ")
	}

	return "Not listed"
}
