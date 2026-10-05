package main

import "net/http"

// disableCaching overrides the Cache-Control header on every response with
// no-store if the -no-cache flag is set, otherwise it does nothing. It needs
// to be the outermost middleware so it can replace headers set by handlers
// and the CacheControl middleware.
func disableCaching() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if !*noCache {
			return next
		}

		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(&noCacheWriter{ResponseWriter: w}, r)
			},
		)
	}
}

type noCacheWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *noCacheWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.wroteHeader = true
		w.Header().Set("Cache-Control", "no-store")
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *noCacheWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func (w *noCacheWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *noCacheWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
