package routing

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"chameth.com/chameth.com/db"
)

type redirect struct {
	matcher     *regexp.Regexp
	destination string
}

var redirects = []redirect{
	{regexp.MustCompile(`^(.*/)index\.html$`), `$1`},
	{regexp.MustCompile(`^(.*)\.html$`), `$1/`},
	{regexp.MustCompile(`^/res/images/sense/(.*)$`), `/sense-api/$1`},
	{regexp.MustCompile(`^/res/images/wemo/(.*)$`), `/monitoring-power-with-wemo/$1`},
	{regexp.MustCompile(`^/res/images/docker/(.*)$`), `/docker-automatic-nginx-proxy/$1`},
	{regexp.MustCompile(`^/res/images/https/(.*)$`), `/why-you-should-be-using-https/$1`},
	{regexp.MustCompile(`^/res/images/yubikey/(.*)$`), `/offline-gnupg-master-yubikey-subkeys/$1`},
	{regexp.MustCompile(`^/res/images/ssh/(.*)$`), `/shoring-up-sshd/$1`},
	{regexp.MustCompile(`^/res/images/android-tests/(.*)$`), `/android-tests-espresso-spoon/$1`},
	{regexp.MustCompile(`^/res/images/dns/(.*)$`), `/top-sites-dns-providers/$1`},
	{regexp.MustCompile(`^/res/images/erl/(.*)$`), `/dns-over-tls-on-edgerouter-lite/$1`},
	{regexp.MustCompile(`^/res/images/aoc/(.*)$`), `/over-the-top-optimisations-in-nim/$1`},
	{regexp.MustCompile(`^/res/images/nim/(.*)$`), `/over-the-top-optimisations-in-nim/$1`},
	{regexp.MustCompile(`^/res/images/debugging/(.*)$`), `/debugging-beyond-the-debugger/$1`},
	{regexp.MustCompile(`^/res/images/unsplash/(.*)$`), `/debugging-beyond-the-debugger/$1`},
	{regexp.MustCompile(`^/res/images/obfuscation/(.*)$`), `/obfuscating-kotlin-proguard/$1`},
	{regexp.MustCompile(`^/20[0-9][0-9]/[0-9][0-9]/[0-9][0-9]/(.*)/?$`), `/$1/`},
	{regexp.MustCompile(`^/page/(.*)/?$`), `/posts/$1/`},
	{regexp.MustCompile(`^/poem/(.*)/?$`), `/$1/`},
	{regexp.MustCompile(`^/prints/([^/]*)/?$`), `/prints/`},
	{regexp.MustCompile(`^/misc/snippets/?$`), `/snippets/`},
	{regexp.MustCompile(`^/feed.xml$`), `/index.xml`},
	{regexp.MustCompile(`^/posts/feed.xml$`), `/index.xml`},
	{regexp.MustCompile(`^/16402FE2.txt$`), `/pgp/`},
	{regexp.MustCompile(`^/favicon.ico`), `/favicon.png`},
	{regexp.MustCompile(`^/misc/?$`), `/sitemap/`},
	{regexp.MustCompile(`^/index/?$`), `/sitemap/`},
	{regexp.MustCompile(`^/about/?$`), `/`},
	{regexp.MustCompile(`^/posts/[0-9]+/?$`), `/posts/`},
	{regexp.MustCompile(`^/catalogue/?$`), `/posts/`},
}

type Manager struct {
	Public *http.ServeMux
	Admin  *http.ServeMux

	contentTypes map[string]http.HandlerFunc
	fallback     http.HandlerFunc
}

func NewManager(fallback http.HandlerFunc) *Manager {
	return &Manager{
		Public: http.NewServeMux(),
		Admin:  http.NewServeMux(),

		contentTypes: make(map[string]http.HandlerFunc),
		fallback:     fallback,
	}
}

func (m *Manager) RegisterContentType(contentType string, handler http.HandlerFunc) {
	m.contentTypes[contentType] = handler
}

func ApplyRedirects() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				for _, rw := range redirects {
					n := rw.matcher.ReplaceAllString(r.URL.Path, rw.destination)
					if n != r.URL.Path {
						// Collapse leading slashes so paths like "//example.com/index.html"
						// can't produce a protocol-relative redirect to another host.
						n = "/" + strings.TrimLeft(n, "/\\")
						w.Header().Add("Location", (&url.URL{Path: n}).EscapedPath())
						w.WriteHeader(http.StatusMovedPermanently)
						return
					}
				}

				next.ServeHTTP(w, r)
			},
		)
	}
}

// IsPage reports whether p is the canonical path of a public page: either
// content in the paths table, or a route registered with a {$} pattern.
// Prefix-matched content and wildcard routes don't count, as they would
// accept arbitrary text in the path.
func (m *Manager) IsPage(ctx context.Context, p string) (bool, error) {
	if canonical := path.Clean(p); p != canonical && p != canonical+"/" {
		return false, nil
	}

	req := &http.Request{Method: http.MethodGet, URL: &url.URL{Path: p}, Header: http.Header{}}
	if _, pattern := m.Public.Handler(req.WithContext(ctx)); strings.HasSuffix(pattern, "{$}") {
		return true, nil
	}

	contentType, err := findExactContentType(ctx, p)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	_, ok := m.contentTypes[contentType]
	return ok, nil
}

func (m *Manager) contentHandler() http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			contentType, err := db.FindContentByPath(r.Context(), r.URL.Path)
			if err != nil {
				slog.Error("Failed to find content by path", "error", err, "path", r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if handler, ok := m.contentTypes[contentType]; ok {
				handler(w, r)
				return
			}

			m.fallback(w, r)
		},
	)
}
