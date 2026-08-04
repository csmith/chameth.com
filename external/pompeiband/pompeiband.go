package pompeiband

import "net/http"

type Client struct {
	h       *http.Client
	baseURL string
}

func NewClient(h *http.Client, baseURL string) *Client {
	return &Client{h: h, baseURL: baseURL}
}
