package feeds

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"net/http"
	"sync"
	"time"
)

const (
	builderCookieName     = "feed-builder"
	builderCookieLifetime = 30 * time.Minute
)

var builderCookieKey = flag.String(
	"feed-builder-cookie-key",
	"",
	"Hex-encoded 32-byte feed builder cookie encryption key; if empty, generate a new key on each startup",
)

var builderCookieCipher = sync.OnceValue(
	func() cipher.AEAD {
		key := make([]byte, 32)
		if *builderCookieKey == "" {
			if _, err := rand.Read(key); err != nil {
				panic(err)
			}
		} else {
			var err error
			key, err = hex.DecodeString(*builderCookieKey)
			if err != nil || len(key) != 32 {
				panic("feed-builder-cookie-key must be a hex-encoded 32-byte key")
			}
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			panic(err)
		}
		aead, err := cipher.NewGCMWithRandomNonce(block)
		if err != nil {
			panic(err)
		}
		return aead
	},
)

func hasBuilderCookie(r *http.Request) bool {
	cookie, err := r.Cookie(builderCookieName)
	if err != nil || len(cookie.Value) > 128 {
		return false
	}
	encoded, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return false
	}
	payload, err := builderCookieCipher().Open(nil, nil, encoded, []byte(builderCookieName))
	if err != nil || len(payload) != 8 {
		return false
	}
	issued := time.Unix(int64(binary.BigEndian.Uint64(payload)), 0)
	now := time.Now()
	return !issued.After(now) && now.Sub(issued) < builderCookieLifetime
}

func setBuilderCookie(w http.ResponseWriter) {
	now := time.Now()
	payload := make([]byte, 8)
	binary.BigEndian.PutUint64(payload, uint64(now.Unix()))
	encoded := builderCookieCipher().Seal(nil, nil, payload, []byte(builderCookieName))
	http.SetCookie(
		w,
		&http.Cookie{
			Name:     builderCookieName,
			Value:    base64.RawURLEncoding.EncodeToString(encoded),
			Path:     builderFeedPrefix,
			Expires:  now.Add(builderCookieLifetime),
			MaxAge:   int(builderCookieLifetime.Seconds()),
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
	)
}

func builderResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.Header().Add("Vary", "Cookie")
}
