// Package servicetoken gets client-credentials tokens that authenticate the
// site to other services.
package servicetoken

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// tokenTimeout bounds each token request, so an unresponsive provider fails
// the service call instead of holding it up.
const tokenTimeout = 10 * time.Second

var client struct {
	issuer       string
	clientID     string
	clientSecret string

	mu       sync.Mutex
	tokenURL string
	sources  map[string]oauth2.TokenSource
}

// Configure sets the OpenID Connect provider and client that tokens are
// requested from. With no client ID or secret, Enabled reports false.
func Configure(issuer, clientID, clientSecret string) {
	client.issuer = issuer
	client.clientID = clientID
	client.clientSecret = clientSecret
}

// Enabled reports whether a client is configured, so callers should send
// tokens.
func Enabled() bool {
	return client.clientID != "" && client.clientSecret != ""
}

// Token returns a token with the given resource as its audience. Tokens are
// cached until they expire.
func Token(ctx context.Context, resource string) (*oauth2.Token, error) {
	source, err := tokenSource(ctx, resource)
	if err != nil {
		return nil, err
	}
	token, err := source.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token for %s: %w", resource, err)
	}
	return token, nil
}

// tokenSource discovers the provider on first use, then keeps one source per
// resource. A failed discovery is not kept, so the next request tries again.
func tokenSource(ctx context.Context, resource string) (oauth2.TokenSource, error) {
	client.mu.Lock()
	defer client.mu.Unlock()

	if source, ok := client.sources[resource]; ok {
		return source, nil
	}

	if client.tokenURL == "" {
		ctx, cancel := context.WithTimeout(ctx, tokenTimeout)
		defer cancel()
		provider, err := oidc.NewProvider(ctx, client.issuer)
		if err != nil {
			return nil, fmt.Errorf("failed to discover OIDC provider: %w", err)
		}
		client.tokenURL = provider.Endpoint().TokenURL
	}

	cfg := clientcredentials.Config{
		ClientID:       client.clientID,
		ClientSecret:   client.clientSecret,
		TokenURL:       client.tokenURL,
		EndpointParams: url.Values{"resource": {resource}},
	}
	// The source outlives the request that creates it, so its token requests
	// get their own context.
	source := cfg.TokenSource(
		context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{Timeout: tokenTimeout}),
	)

	if client.sources == nil {
		client.sources = map[string]oauth2.TokenSource{}
	}
	client.sources[resource] = source
	return source, nil
}
