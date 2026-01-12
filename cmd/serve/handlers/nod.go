package handlers

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

var (
	ircCatAddress    = flag.String("irc-cat-address", "", "URL to post IRC notifications")
	ircCatApiKey     = flag.String("irc-cat-key", "", "API key for IRC notifications")
	ircCatNodChannel = flag.String("irc-cat-nod-channel", "", "Channel to post nod messages to")
)

type nodRequest struct {
	Page string `json:"page"`
}

func Nod(w http.ResponseWriter, r *http.Request) {
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

	var nr nodRequest
	if err = json.Unmarshal(body, &nr); err != nil {
		slog.Error("Error parsing nod payload", "error", err, "payload", string(body))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(nr.Page) == 0 || len(nr.Page) > 512 {
		slog.Error("Invalid page length for nod", "page", nr.Page, "length", len(nr.Page))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	parsedURL, err := url.Parse(nr.Page)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		slog.Error("Invalid URL format for nod", "page", nr.Page, "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(parsedURL.Scheme, "http") {
		slog.Error("Invalid URL scheme for nod", "page", nr.Page, "scheme", parsedURL.Scheme)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err = announceToIrc(fmt.Sprintf("\00311\002[CHAMETH.COM]\002\003 Someone nodded at %s", nr.Page)); err != nil {
		slog.Error("Error announcing nod messages to IRC", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func announceToIrc(message string) error {
	type Wrapper struct {
		Message string `json:"message"`
		Channel string `json:"channel"`
	}

	b, err := json.Marshal(Wrapper{message, *ircCatNodChannel})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, *ircCatAddress, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", *ircCatApiKey)
	r, err := http.DefaultClient.Do(req)
	_ = r.Body.Close()
	return err
}
