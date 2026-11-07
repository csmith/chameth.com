// Package twowords fetches data from the personal services on the tailnet.
package twowords

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

// ErrNotFound is returned when the service responds with a 404.
var ErrNotFound = errors.New("not found")

// Service is a tailnet HTTP service; Name is only used in log messages.
type Service struct {
	Name    string
	BaseURL string
}

// Query holds request parameters. Values are formatted with fmt.Sprint, and
// zero values ("", 0, false, nil) are omitted.
type Query map[string]any

// Get fetches path (which should be root-relative, see Path) with the given
// query, returning the response body and headers.
func (s Service) Get(ctx context.Context, client *http.Client, path string, query Query) ([]byte, http.Header, error) {
	u := strings.TrimRight(s.BaseURL, "/") + path
	if values := query.values(); len(values) > 0 {
		u += "?" + values.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build request: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch %s: %w", path, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %s response: %w", path, err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, nil, ErrNotFound
	}
	if res.StatusCode >= 400 {
		slog.Warn(s.Name+" request failed", "path", path, "status", res.StatusCode, "response", string(body))
		return nil, nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}
	return body, res.Header, nil
}

// Path formats a request path, path-escaping any string arguments.
func Path(format string, args ...any) string {
	for i, arg := range args {
		if s, ok := arg.(string); ok {
			args[i] = url.PathEscape(s)
		}
	}
	return fmt.Sprintf(format, args...)
}

func (q Query) values() url.Values {
	values := url.Values{}
	for k, v := range q {
		if v == nil || reflect.ValueOf(v).IsZero() {
			continue
		}
		values.Set(k, fmt.Sprint(v))
	}
	return values
}
