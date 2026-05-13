package nod

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func HandleJSON(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Error reading nod body", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var req struct {
		Page string `json:"page"`
	}
	if err = json.Unmarshal(body, &req); err != nil {
		slog.Error("Error parsing nod payload", "error", err, "payload", string(body))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err = processNod(req.Page); err != nil {
		slog.Error("Error processing nod", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func HandleForm(w http.ResponseWriter, r *http.Request) {
	if err := processNod(r.FormValue("page")); err != nil {
		slog.Error("Error processing nod form", "error", err)
		http.Error(w, "Something went wrong", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "Your nod has been received and is appreciated")
}
