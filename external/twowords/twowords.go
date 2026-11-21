// Package twowords fetches data from the personal services on the tailnet.
package twowords

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"chameth.com/x/oidc"
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

// GetJSON fetches path (which must be root-relative, see Path) with the given
// query, and decodes the JSON response into a T.
func GetJSON[T any](ctx context.Context, s Service, client *http.Client, path string, query Query) (T, error) {
	var res T
	body, _, err := s.get(ctx, client, path, query)
	if err != nil {
		return res, err
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return res, fmt.Errorf("failed to decode %s response: %w", path, err)
	}
	return res, nil
}

// GetImage fetches the image at path (which must be root-relative), returning
// its body and Content-Type.
func (s Service) GetImage(ctx context.Context, client *http.Client, path string) ([]byte, string, error) {
	body, header, err := s.get(ctx, client, path, nil)
	if err != nil {
		return nil, "", err
	}
	return body, header.Get("Content-Type"), nil
}

// Tokens, when set, gets the access token sent with each request, with the
// service's base URL as the resource.
var Tokens *oidc.ClientCredentials

func (s Service) get(ctx context.Context, client *http.Client, path string, query Query) ([]byte, http.Header, error) {
	// Paths can come from earlier responses; anything not root-relative
	// could move the request, and the token, to another host or port.
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return nil, nil, fmt.Errorf("path %q is not root-relative", path)
	}

	u := strings.TrimRight(s.BaseURL, "/") + path
	if values := query.values(); len(values) > 0 {
		u += "?" + values.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build request: %w", err)
	}
	if Tokens != nil {
		token, err := Tokens.Token(ctx, s.BaseURL)
		if err != nil {
			return nil, nil, err
		}
		token.SetAuthHeader(req)
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
