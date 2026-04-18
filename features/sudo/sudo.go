package sudo

import (
	"context"
	"net/http"
)

type contextKey int

var adminKey contextKey

func IsAdmin(ctx context.Context) bool {
	admin, _ := ctx.Value(adminKey).(bool)
	return admin
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("isadmin"); err == nil && cookie.Value == "1" {
			r = r.WithContext(context.WithValue(r.Context(), adminKey, true))
		}
		next.ServeHTTP(w, r)
	})
}

func SetCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "isadmin",
		Value:    "1",
		MaxAge:   86400 * 365 * 10,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
