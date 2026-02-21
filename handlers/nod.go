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

func processNod(page string) error {
	if len(page) == 0 || len(page) > 512 {
		return fmt.Errorf("invalid page length: %d", len(page))
	}

	parsedURL, err := url.Parse(page)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("invalid URL format: %s", page)
	}

	if !strings.HasPrefix(parsedURL.Scheme, "http") {
		return fmt.Errorf("invalid URL scheme: %s", parsedURL.Scheme)
	}

	return announceToIrc(fmt.Sprintf("\00311\002[CHAMETH.COM]\002\003 Someone nodded at %s", page))
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

	if err = processNod(nr.Page); err != nil {
		slog.Error("Error processing nod", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func NodForm(w http.ResponseWriter, r *http.Request) {
	if err := processNod(r.FormValue("page")); err != nil {
		slog.Error("Error processing nod form", "error", err)
		http.Error(w, "Something went wrong", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "Your nod has been received and is appreciated")
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
