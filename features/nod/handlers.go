package nod

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"go.temporal.io/sdk/client"
)

const maxBodySize = 4096

func handleJSON(pages pageChecker, tc client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		}

		defer r.Body.Close()
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodySize))
		if err != nil {
			slog.Info("Error reading nod body", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var req struct {
			Page string `json:"page"`
		}
		if err = json.Unmarshal(body, &req); err != nil {
			slog.Info("Error parsing nod payload", "error", err, "size", len(body))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err = processNod(r.Context(), pages, tc, r.RemoteAddr, req.Page); err != nil {
			status := nodErrorStatus(err)
			http.Error(w, http.StatusText(status), status)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func handleForm(pages pageChecker, tc client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
		if err := r.ParseForm(); err != nil {
			slog.Info("Error parsing nod form", "error", err)
			http.Error(w, "Something went wrong", http.StatusBadRequest)
			return
		}

		if err := processNod(r.Context(), pages, tc, r.RemoteAddr, r.FormValue("page")); err != nil {
			http.Error(w, "Something went wrong", nodErrorStatus(err))
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, "Your nod has been received and is appreciated")
	}
}

// nodErrorStatus logs a processNod error and returns the status to respond
// with: 400 if the client sent a bad page, 500 otherwise.
func nodErrorStatus(err error) int {
	if _, ok := errors.AsType[*invalidNodError](err); ok {
		slog.Info("Rejecting invalid nod", "error", err)
		return http.StatusBadRequest
	}
	slog.Error("Error processing nod", "error", err)
	return http.StatusInternalServerError
}
