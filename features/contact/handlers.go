package contact

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"go.temporal.io/sdk/client"
)

func handleJSON(tc client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		if err := startWorkflow(r.Context(), tc, newSubmission(r, req, methodJSON)); err != nil {
			slog.Error("Error starting contact workflow", "error", err, "request", req)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func handleForm(tc client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := request{
			Page:        r.FormValue("page"),
			SenderName:  r.FormValue("name"),
			SenderEmail: r.FormValue("email"),
			Message:     r.FormValue("message"),
			Timestamp:   r.FormValue("ts"),
			Honeypot:    r.FormValue("subject"),
		}

		if err := startWorkflow(r.Context(), tc, newSubmission(r, req, methodForm)); err != nil {
			slog.Error("Error starting contact workflow", "error", err, "request", req)
			http.Error(w, "Something went wrong. Your message was not sent.", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, "Message sent! Thanks for getting in touch!")
	}
}

func newSubmission(r *http.Request, req request, mthd method) submission {
	return submission{
		Request:    req,
		Method:     mthd,
		RemoteAddr: r.RemoteAddr,
		UserAgent:  r.UserAgent(),
		ReceivedAt: time.Now(),
	}
}
