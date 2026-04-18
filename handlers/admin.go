package handlers

import (
	"net/http"

	"chameth.com/chameth.com/features/sudo"
)

func Sudo(w http.ResponseWriter, r *http.Request) {
	sudo.SetCookie(w)
	http.Redirect(w, r, "/", http.StatusFound)
}
