package twowords

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

var auth struct {
	issuer       string
	clientID     string
	clientSecret string

	mu       sync.Mutex
	tokenURL string
	sources  map[string]oauth2.TokenSource
}

// UseClient makes requests carry a client-credentials token from the given
// OpenID Connect provider, with the service's base URL as its audience. With
// no client ID or secret, requests are sent without a token.
func UseClient(issuer, clientID, clientSecret string) {
	auth.issuer = issuer
	auth.clientID = clientID
	auth.clientSecret = clientSecret
}

func (s Service) authorise(ctx context.Context, req *http.Request) error {
	if auth.clientID == "" || auth.clientSecret == "" {
		return nil
	}

	source, err := s.tokenSource(ctx)
	if err != nil {
		return err
	}
	token, err := source.Token()
	if err != nil {
		return fmt.Errorf("failed to get token for %s: %w", s.Name, err)
	}
	token.SetAuthHeader(req)
	return nil
}

// tokenSource discovers the provider on first use, then keeps one source per
// service, which caches the token until it expires. A failed discovery is
// not kept, so the next request tries again.
func (s Service) tokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	auth.mu.Lock()
	defer auth.mu.Unlock()

	if source, ok := auth.sources[s.BaseURL]; ok {
		return source, nil
	}

	if auth.tokenURL == "" {
		ctx, cancel := context.WithTimeout(ctx, tokenTimeout)
		defer cancel()
		provider, err := oidc.NewProvider(ctx, auth.issuer)
		if err != nil {
			return nil, fmt.Errorf("failed to discover OIDC provider: %w", err)
		}
		auth.tokenURL = provider.Endpoint().TokenURL
	}

	cfg := clientcredentials.Config{
		ClientID:       auth.clientID,
		ClientSecret:   auth.clientSecret,
		TokenURL:       auth.tokenURL,
		EndpointParams: url.Values{"resource": {s.BaseURL}},
	}
	// The source outlives the request that creates it, so its token requests
	// get their own context.
	source := cfg.TokenSource(
		context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{Timeout: tokenTimeout}),
	)

	if auth.sources == nil {
		auth.sources = map[string]oauth2.TokenSource{}
	}
	auth.sources[s.BaseURL] = source
	return source, nil
}
