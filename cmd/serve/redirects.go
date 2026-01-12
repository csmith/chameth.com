package main

import (
	"net/http"
	"regexp"
)

type redirect struct {
	matcher     *regexp.Regexp
	destination string
}

var redirects = []redirect{
	// URLs should be folders, not individual HTML files
	{regexp.MustCompile(`^(.*/)index\.html$`), `$1`},
	{regexp.MustCompile(`^(.*)\.html$`), `$1/`},
	// Old paths for images from before "content bundles"
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
	// Old paths for posts and other bits
	{regexp.MustCompile(`^/20[0-9][0-9]/[0-9][0-9]/[0-9][0-9]/(.*)/?$`), `/$1/`},
	{regexp.MustCompile(`^/page/(.*)/?$`), `/posts/$1/`},
	{regexp.MustCompile(`^/poem/(.*)/?$`), `/$1/`},
	{regexp.MustCompile(`^/prints/(.*)/?$`), `/prints/`},
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

func applyRedirects() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, rw := range redirects {
				n := rw.matcher.ReplaceAllString(r.URL.Path, rw.destination)
				if n != r.URL.Path {
					w.Header().Add("Location", n)
					w.WriteHeader(http.StatusMovedPermanently)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
