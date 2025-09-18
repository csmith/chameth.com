package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/smtp"
	"strings"
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
	body.WriteString("SOURCE IP: ")
	body.WriteString(req.RemoteAddr)
	body.WriteString("\n\n")
	body.WriteString("MESSAGE:\n\n")
	body.WriteString(c.Message)
	return body.String()
}
