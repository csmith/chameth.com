package metrics

import (
	"net/http"
	"time"
)

// LogRequests records every request, including responses short-circuited by outer middleware.
func LogRequests() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			cw := &outerWriter{ResponseWriter: w}
			next.ServeHTTP(cw, r)
			ip := remoteAddress(r.RemoteAddr)
			size := cw.n
			if r.Method == http.MethodHead {
				size = 0
			}
			enqueueRequestLog(requestLog{url: truncateString(r.URL.RequestURI(), maxLoggedURLLength), userAgent: truncateString(r.UserAgent(), maxLoggedUserAgentLength), ip: ip, ipHash: hashIP(ip, start), start: start, duration: time.Since(start), size: size, status: cw.status()})
		})
	}
}

type outerWriter struct {
	http.ResponseWriter
	code, n int
}

func (w *outerWriter) WriteHeader(code int) {
	if code >= 100 && code < 200 {
		w.ResponseWriter.WriteHeader(code)
		return
	}
	if w.code != 0 {
		return
	}
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}
func (w *outerWriter) Write(b []byte) (int, error) {
	if w.code == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, e := w.ResponseWriter.Write(b)
	w.n += n
	return n, e
}
func (w *outerWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		if w.code == 0 {
			w.WriteHeader(http.StatusOK)
		}
		f.Flush()
	}
}
func (w *outerWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func (w *outerWriter) status() int {
	if w.code == 0 {
		return http.StatusOK
	}
	return w.code
}
