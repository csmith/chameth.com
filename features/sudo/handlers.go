package sudo

import (
	"net/http"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	SetCookie(w)
	http.Redirect(w, r, "/", http.StatusFound)
}
