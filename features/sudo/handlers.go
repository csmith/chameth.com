package sudo

import (
	"net/http"
)

func handle(w http.ResponseWriter, r *http.Request) {
	SetCookie(w)
	http.Redirect(w, r, "/", http.StatusFound)
}
