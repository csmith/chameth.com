package admin

import (
	"context"
	"flag"
	"log/slog"
	"net/http"

	"chameth.com/x/oidc"
)

var (
	oidcIssuer = flag.String(
		"oidc-issuer",
		"https://id.929w.net",
		"OpenID Connect issuer for admin logins and service calls",
	)
	oidcClientID     = flag.String("oidc-client-id", "", "OpenID Connect client ID for admin logins and service calls")
	oidcClientSecret = flag.String(
		"oidc-client-secret",
		"",
		"OpenID Connect client secret for admin logins and service calls",
	)
	oidcSessionKey = flag.String(
		"oidc-session-key",
		"",
		"Key that signs admin session cookies, at least 32 characters; changing it logs everyone out",
	)
	noLogin = flag.Bool(
		"no-login",
		false,
		"Serve the admin interface without requiring a login, e.g. for local development",
	)
)

// OIDCClient returns the OpenID Connect client that admin logins use, which
// the site also uses to authenticate to other services.
func OIDCClient() (issuer, clientID, clientSecret string) {
	return *oidcIssuer, *oidcClientID, *oidcClientSecret
}

// requireLogin puts every page served by next behind a login, using the
// -oidc-* flags. With -no-login it returns next unchanged.
func requireLogin(next http.Handler, baseURL func(context.Context) (string, error)) (http.Handler, error) {
	if *noLogin {
		slog.Warn("Serving the admin interface without a login")
		return next, nil
	}
	login, err := oidc.NewBrowserLogin(
		oidc.BrowserLoginConfig{
			Issuer:       *oidcIssuer,
			ClientID:     *oidcClientID,
			ClientSecret: *oidcClientSecret,
			SessionKey:   *oidcSessionKey,
			BaseURLFunc:  baseURL,
		},
	)
	if err != nil {
		return nil, err
	}
	return login.Wrap(next), nil
}
