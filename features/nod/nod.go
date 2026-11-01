package nod

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"chameth.com/chameth.com/templates"
)

var (
	ircCatAddress    = flag.String("irc-cat-address", "", "URL to post IRC notifications")
	ircCatApiKey     = flag.String("irc-cat-key", "", "API key for IRC notifications")
	ircCatNodChannel = flag.String("irc-cat-nod-channel", "", "Channel to post nod messages to")

	httpClient = &http.Client{Timeout: 10 * time.Second}
)

type pageChecker interface {
	IsPage(ctx context.Context, path string) (bool, error)
}

// processNod announces a nod for page, which may be a path or a full URL;
// only its path is used, and only if it is a page on this site.
func processNod(ctx context.Context, pages pageChecker, remoteAddr, page string) error {
	if len(page) == 0 || len(page) > 512 {
		return fmt.Errorf("invalid page length: %d", len(page))
	}

	parsedURL, err := url.Parse(page)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	path := parsedURL.Path
	if ok, err := pages.IsPage(ctx, path); err != nil {
		return fmt.Errorf("unable to check page %q: %w", path, err)
	} else if !ok {
		return fmt.Errorf("not a page: %q", path)
	}

	if !nodAllowed(remoteAddr, time.Now()) {
		slog.Info("Dropping nod over rate limit", "remoteAddr", remoteAddr, "path", path)
		return nil
	}

	return announceToIrc(
		"\00311\002[CHAMETH.COM]\002\003 Someone nodded at " +
			templates.SiteURL() + (&url.URL{Path: path}).EscapedPath(),
	)
}

const nodInterval = time.Minute

var (
	nodsMu   sync.Mutex
	lastNods = make(map[string]time.Time)
)

// nodAllowed reports whether the client at remoteAddr may nod now, allowing
// one nod per client per nodInterval.
func nodAllowed(remoteAddr string, now time.Time) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	nodsMu.Lock()
	defer nodsMu.Unlock()

	for h, t := range lastNods {
		if now.Sub(t) >= nodInterval {
			delete(lastNods, h)
		}
	}

	if _, ok := lastNods[host]; ok {
		return false
	}
	lastNods[host] = now
	return true
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

	// The IRC announce is fire-and-forget: it must not be cancelled if the
	// requesting client disconnects, so it uses a background context.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, *ircCatAddress, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", *ircCatApiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to irc-cat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("irc-cat returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}
