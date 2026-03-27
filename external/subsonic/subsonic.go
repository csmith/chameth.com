package subsonic

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/url"
)

type Client struct {
	h        *http.Client
	baseURL  string
	username string
	password string
}

func NewClient(h *http.Client, baseURL, username, password string) *Client {
	return &Client{
		h:        h,
		baseURL:  baseURL,
		username: username,
		password: password,
	}
}

func (c *Client) url(endpoint string, params url.Values) string {
	if params == nil {
		params = url.Values{}
	}
	params.Set("u", c.username)
	params.Set("p", c.password)
	params.Set("v", "1.16.1")
	params.Set("c", "chameth.com")
	return fmt.Sprintf("%s/rest/%s?%s", c.baseURL, endpoint, params.Encode())
}

func (c *Client) get(endpoint, key string, result any, extraParams url.Values) error {
	params := url.Values{"f": {"json"}}
	maps.Copy(params, extraParams)
	u := c.url(endpoint, params)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}

	res, err := c.h.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		slog.Warn("Subsonic request failed", "status", res.StatusCode, "endpoint", endpoint, "response", string(b))
		return fmt.Errorf("bad status code: %d", res.StatusCode)
	}

	var envelope struct {
		Fields map[string]json.RawMessage `json:"subsonic-response"`
	}
	if err := json.Unmarshal(b, &envelope); err != nil {
		return err
	}

	var check struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if content, ok := envelope.Fields["error"]; ok {
		if err := json.Unmarshal(content, &check); err != nil {
			return err
		}
		if check.Code != 0 {
			return fmt.Errorf("subsonic error %d: %s", check.Code, check.Message)
		}
	}

	raw, ok := envelope.Fields[key]
	if !ok {
		return fmt.Errorf("subsonic response missing %q field", key)
	}

	return json.Unmarshal(raw, result)
}
