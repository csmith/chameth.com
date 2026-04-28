package contact

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

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

	var req request
	if err = json.Unmarshal(body, &req); err != nil {
		slog.Error("Error parsing contact form payload", "error", err, "payload", string(body))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := process(r.Context(), req, methodJSON, r.RemoteAddr, r.UserAgent()); err != nil {
		if errors.Is(err, errRejected) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func ContactFormPost(w http.ResponseWriter, r *http.Request) {
	req := request{
		Page:        r.FormValue("page"),
		SenderName:  r.FormValue("name"),
		SenderEmail: r.FormValue("email"),
		Message:     r.FormValue("message"),
		Timestamp:   r.FormValue("ts"),
		Honeypot:    r.FormValue("subject"),
	}

	if err := process(r.Context(), req, methodForm, r.RemoteAddr, r.UserAgent()); err != nil {
		if errors.Is(err, errRejected) {
			http.Error(w, "Something went wrong. Your message was not sent.", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong. Your message was not sent.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "Message sent! Thanks for getting in touch!")
}
