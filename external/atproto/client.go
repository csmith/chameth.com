package atproto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Client struct {
	h            *http.Client
	pds          pds
	handle       string
	did          string
	accessToken  string
	refreshToken string
}

func NewClient(pdsUrl, handle, password string) (*Client, error) {
	client := &Client{
		h: &http.Client{
			Timeout: 30 * time.Second,
		},
		pds:    pds(pdsUrl),
		handle: handle,
	}

	if err := client.authenticate(password); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) authenticate(password string) error {
	payload := struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}{
		Identifier: c.handle,
		Password:   password,
	}

	result := struct {
		AccessJWT  string `json:"accessJwt"`
		RefreshJWT string `json:"refreshJwt"`
		Handle     string `json:"handle"`
		DID        string `json:"did"`
	}{}

	if err := c.postJson(createSessionEndpoint, payload, &result); err != nil {
		return err
	}

	c.accessToken = result.AccessJWT
	c.refreshToken = result.RefreshJWT
	c.did = result.DID
	return nil
}

func (c *Client) DID() string {
	return c.did
}

func (c *Client) CreateRecord(collection Collection, record Record) (StrongRef, string, error) {
	recordKey := generateTID()

	payload := struct {
		Repository string     `json:"repo"`
		Collection Collection `json:"collection"`
		RecordKey  string     `json:"rkey"`
		Record     Record     `json:"record"`
	}{
		Repository: c.did,
		Collection: collection,
		RecordKey:  recordKey,
		Record:     record,
	}

	result := struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}{}

	if err := c.postJson(putRecordEndpoint, payload, &result); err != nil {
		return StrongRef{}, "", err
	}

	return StrongRef{CID: result.CID, URI: result.URI}, collection.publicURL(c.handle, recordKey), nil
}

func (c *Client) GetRecord(collection Collection, recordKey string) (StrongRef, error) {
	e := endpoint(fmt.Sprintf("%s?repo=%s&collection=%s&rkey=%s", getRecordEndpoint, c.did, collection, recordKey))

	req, err := http.NewRequest(http.MethodGet, c.pds.url(e), nil)
	if err != nil {
		return StrongRef{}, err
	}

	req.Header.Set("Accept", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	}

	res, err := c.h.Do(req)
	if err != nil {
		return StrongRef{}, err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return StrongRef{}, err
	}

	if res.StatusCode >= 400 {
		return StrongRef{}, fmt.Errorf("bad status code: %d: %s", res.StatusCode, string(b))
	}

	var result struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return StrongRef{}, err
	}

	return StrongRef{CID: result.CID, URI: result.URI}, nil
}

func (c *Client) UploadBlob(mimeType string, data []byte) (*Blob, error) {
	var result struct {
		Blob Blob `json:"blob"`
	}
	if err := c.post(uploadBlobEndpoint, mimeType, bytes.NewReader(data), &result); err != nil {
		return nil, err
	}
	return &result.Blob, nil
}

func (c *Client) postJson(endpoint endpoint, payload any, result any) error {
	marshalled, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.post(endpoint, "application/json", bytes.NewReader(marshalled), &result)
}

func (c *Client) post(endpoint endpoint, contentType string, payload io.Reader, result any) error {
	req, err := http.NewRequest(http.MethodPost, c.pds.url(endpoint), payload)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
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
		slog.Warn("PDS request failed", "status", res.StatusCode, "endpoint", endpoint, "response", string(b))
		return fmt.Errorf("bad status code: %d", res.StatusCode)
	}

	return json.Unmarshal(b, &result)
}
