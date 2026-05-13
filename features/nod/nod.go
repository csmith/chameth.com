package nod

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var (
	ircCatAddress    = flag.String("irc-cat-address", "", "URL to post IRC notifications")
	ircCatApiKey     = flag.String("irc-cat-key", "", "API key for IRC notifications")
	ircCatNodChannel = flag.String("irc-cat-nod-channel", "", "Channel to post nod messages to")
)

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
