package nod

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"sync"
	"time"

	"chameth.com/chameth.com/templates"
	"go.temporal.io/sdk/client"
)

const (
	ircTaskQueue    = "irc"
	ircWorkflowName = "irc.SendMessage"
	ircSource       = "chameth.com"
	ircStartTimeout = 10 * time.Second
)

type pageChecker interface {
	IsPage(ctx context.Context, path string) (bool, error)
}

// processNod announces a nod for page, which may be a path or a full URL;
// only its path is used, and only if it is a page on this site.
func processNod(ctx context.Context, pages pageChecker, tc client.Client, remoteAddr, page string) error {
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

	return announceToIrc(tc, "Someone nodded at "+templates.SiteURL()+(&url.URL{Path: path}).EscapedPath())
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

func announceToIrc(tc client.Client, message string) error {
	type input struct {
		Source  string `json:"source"`
		Message string `json:"message"`
	}

	// The IRC announce is fire-and-forget: it must not be cancelled if the
	// requesting client disconnects, so it uses a background context.
	ctx, cancel := context.WithTimeout(context.Background(), ircStartTimeout)
	defer cancel()

	_, err := tc.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{TaskQueue: ircTaskQueue},
		ircWorkflowName,
		input{Source: ircSource, Message: message},
	)
	if err != nil {
		return fmt.Errorf("failed to start IRC announcement: %w", err)
	}
	return nil
}
