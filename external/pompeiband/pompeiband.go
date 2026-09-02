package pompeiband

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

// BaseURL is the Pompei Band daemon's address on the tailnet. The client
// routes requests over an injected *http.Client (typically the tsnet
// server's); dialing brings the tailnet up if it isn't already connected.
const BaseURL = "https://pb.yak-wall.ts.net"

type Client struct {
	h *http.Client
}

func NewClient(h *http.Client) *Client {
	return &Client{h: h}
}

// get performs a GET request against the API, returning the response body.
// A 404 is reported as ErrNotFound; any other error status is logged and
// returned as an error.
func (c *Client) get(ctx context.Context, path string, query url.Values) ([]byte, error) {
	u := strings.TrimRight(BaseURL, "/") + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	res, err := c.h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", path, err)
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s response: %w", path, err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if res.StatusCode >= 400 {
		slog.Warn("Pompei Band request failed", "path", path, "status", res.StatusCode, "response", string(b))
		return nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}
	return b, nil
}
