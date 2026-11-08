package admin

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
	_ "time/tzdata" // the session reset is in Europe/London, and containers often ship without zoneinfo

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	oidcIssuer = flag.String(
		"oidc-issuer",
		"https://id.929w.net",
		"OpenID Connect issuer that admin logins go through",
	)
	oidcClientID     = flag.String("oidc-client-id", "", "OpenID Connect client ID for admin logins")
	oidcClientSecret = flag.String("oidc-client-secret", "", "OpenID Connect client secret for admin logins")
	oidcSessionKey   = flag.String(
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

const (
	callbackPath      = "/auth/callback"
	sessionCookie     = "auth_session"
	stateCookiePrefix = "auth_state_"
	loginTimeout      = 10 * time.Minute
	minKeyLength      = 32
)

// requireLogin puts every page served by next behind a login, using the
// -oidc-* flags. With -no-login it returns next unchanged. Provider discovery
// happens on the first login, so the admin interface still starts if the
// provider is down.
func requireLogin(next http.Handler, baseURL func(context.Context) (string, error)) (http.Handler, error) {
	if *noLogin {
		slog.Warn("Serving the admin interface without a login")
		return next, nil
	}
	if *oidcClientID == "" || *oidcClientSecret == "" {
		return nil, errors.New("oidc-client-id and oidc-client-secret must both be set to require a login")
	}
	if len(*oidcSessionKey) < minKeyLength {
		return nil, fmt.Errorf("oidc-session-key must be at least %d characters", minKeyLength)
	}
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		return nil, fmt.Errorf("load session reset time zone: %w", err)
	}

	l := &login{
		issuer:       *oidcIssuer,
		clientID:     *oidcClientID,
		clientSecret: *oidcClientSecret,
		key:          []byte(*oidcSessionKey),
		london:       london,
		baseURL:      baseURL,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET "+callbackPath, l.handleCallback)
	mux.Handle("/", l.middleware(next))
	return mux, nil
}

type login struct {
	issuer       string
	clientID     string
	clientSecret string
	key          []byte
	london       *time.Location
	baseURL      func(context.Context) (string, error)

	mu       sync.Mutex
	oauth    *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func (l *login) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if l.hasSession(r) {
				next.ServeHTTP(w, r)
				return
			}
			// Only a navigation can usefully be sent off to log in. Anything else
			// would arrive back at the page as a GET, without its body.
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				http.Error(w, "Not logged in", http.StatusUnauthorized)
				return
			}
			l.redirectToProvider(w, r)
		},
	)
}

type session struct {
	Subject string    `json:"sub"`
	Expires time.Time `json:"exp"`
}

func (l *login) hasSession(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return false
	}
	var s session
	return l.open(sessionCookie, cookie.Value, &s) && time.Now().Before(s.Expires)
}

// pending is what the callback needs to finish a login. It is kept in a cookie
// named after the state, so several tabs sent to log in at once each find
// their own.
type pending struct {
	Verifier string    `json:"verifier"`
	Nonce    string    `json:"nonce"`
	Return   string    `json:"return"`
	Expires  time.Time `json:"exp"`
}

func (l *login) redirectToProvider(w http.ResponseWriter, r *http.Request) {
	cfg, _, err := l.client(r.Context())
	if err != nil {
		slog.Error("Failed to set up login", "error", err)
		http.Error(w, "Login is unavailable", http.StatusBadGateway)
		return
	}

	state := rand.Text()
	p := pending{
		Verifier: oauth2.GenerateVerifier(),
		Nonce:    rand.Text(),
		Return:   r.URL.RequestURI(),
		Expires:  time.Now().Add(loginTimeout),
	}
	name := stateCookiePrefix + state
	value, err := l.seal(name, p)
	if err != nil {
		slog.Error("Failed to start login", "error", err)
		http.Error(w, "Login is unavailable", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, l.cookie(name, value, callbackPath, p.Expires))
	http.Redirect(
		w,
		r,
		cfg.AuthCodeURL(state, oauth2.S256ChallengeOption(p.Verifier), oidc.Nonce(p.Nonce)),
		http.StatusFound,
	)
}

func (l *login) handleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	name := stateCookiePrefix + query.Get("state")
	cookie, err := r.Cookie(name)
	var p pending
	if query.Get("state") == "" || err != nil || !l.open(name, cookie.Value, &p) || !time.Now().Before(p.Expires) {
		http.Error(w, "This login has expired; go back to the page and try again", http.StatusBadRequest)
		return
	}

	cfg, verifier, err := l.client(r.Context())
	if err != nil {
		slog.Error("Failed to set up login", "error", err)
		http.Error(w, "Login is unavailable", http.StatusBadGateway)
		return
	}
	http.SetCookie(w, l.cookie(name, "", callbackPath, time.Unix(0, 0)))

	if e := query.Get("error"); e != "" {
		slog.Warn("Login refused by provider", "error", e, "description", query.Get("error_description"))
		http.Error(w, "Login was refused", http.StatusForbidden)
		return
	}

	token, err := cfg.Exchange(r.Context(), query.Get("code"), oauth2.VerifierOption(p.Verifier))
	if err != nil {
		slog.Error("Failed to exchange login code", "error", err)
		http.Error(w, "Login failed", http.StatusBadGateway)
		return
	}
	raw, ok := token.Extra("id_token").(string)
	if !ok {
		slog.Error("Login token response had no ID token")
		http.Error(w, "Login failed", http.StatusBadGateway)
		return
	}
	id, err := verifier.Verify(r.Context(), raw)
	if err != nil || id.Nonce != p.Nonce {
		slog.Warn("Rejected login ID token", "error", err)
		http.Error(w, "Login failed", http.StatusForbidden)
		return
	}

	s := session{Subject: id.Subject, Expires: nextReset(time.Now(), l.london)}
	value, err := l.seal(sessionCookie, s)
	if err != nil {
		slog.Error("Failed to create session", "error", err)
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, l.cookie(sessionCookie, value, "/", s.Expires))
	http.Redirect(w, r, localPath(p.Return), http.StatusFound)
}

// client discovers the provider on first use and keeps the result. A failure
// is not kept, so the next request tries again. The callback URL is worked
// out at the same point because it needs the tailnet to be up.
func (l *login) client(ctx context.Context) (*oauth2.Config, *oidc.IDTokenVerifier, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.oauth != nil {
		return l.oauth, l.verifier, nil
	}

	base, err := l.baseURL(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("work out callback URL: %w", err)
	}
	provider, err := oidc.NewProvider(ctx, l.issuer)
	if err != nil {
		return nil, nil, fmt.Errorf("discover OIDC provider: %w", err)
	}
	l.oauth = &oauth2.Config{
		ClientID:     l.clientID,
		ClientSecret: l.clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  base + callbackPath,
		Scopes:       []string{oidc.ScopeOpenID},
	}
	l.verifier = provider.Verifier(&oidc.Config{ClientID: l.clientID})
	return l.oauth, l.verifier, nil
}

func (l *login) cookie(name, value, path string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Expires:  expires,
		HttpOnly: true,
		Secure:   true,
		// Lax rather than Strict: the provider's redirect back to the callback is a
		// cross-site navigation, and it has to carry the state cookie.
		SameSite: http.SameSiteLaxMode,
	}
}

// seal signs v under the cookie's name as well as the key, so a valid value
// for one cookie is not accepted for another.
func (l *login) seal(name string, v any) (string, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return encoded + "." + base64.RawURLEncoding.EncodeToString(l.mac(name, encoded)), nil
}

func (l *login) open(name, value string, v any) bool {
	encoded, sig, ok := strings.Cut(value, ".")
	if !ok {
		return false
	}
	mac, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil || !hmac.Equal(mac, l.mac(name, encoded)) {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return false
	}
	return json.Unmarshal(payload, v) == nil
}

func (l *login) mac(name, encoded string) []byte {
	h := hmac.New(sha256.New, l.key)
	h.Write([]byte(name))
	h.Write([]byte{0})
	h.Write([]byte(encoded))
	return h.Sum(nil)
}

// nextReset is the first 05:00 in loc strictly after t. Both instants are
// built with time.Date, so it is still 05:00 on the days the clocks change.
func nextReset(t time.Time, loc *time.Location) time.Time {
	local := t.In(loc)
	reset := time.Date(local.Year(), local.Month(), local.Day(), 5, 0, 0, 0, loc)
	if !reset.After(local) {
		reset = time.Date(local.Year(), local.Month(), local.Day()+1, 5, 0, 0, 0, loc)
	}
	return reset
}

// localPath keeps the post-login redirect on the admin host. The stored value
// is signed, so this is a second check rather than the only one.
func localPath(p string) string {
	if !strings.HasPrefix(p, "/") || strings.HasPrefix(p, "//") || strings.HasPrefix(p, `/\`) {
		return "/"
	}
	return p
}
