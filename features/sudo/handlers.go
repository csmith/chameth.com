package sudo

import (
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	SetCookie(w)
	http.Redirect(w, r, "/", http.StatusFound)
}
