package metrics

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/csmith/aca"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type contextKey int

var (
	requestIdKey contextKey

	inFlightRequestsMu sync.RWMutex
	inFlightRequests   = make(map[string]*request)

	buildVersion string
)

type request struct {
	start   time.Time
	queries atomic.Int32
}

func normalizePath(path string) string {
	// Limit label cardinality a bit by truncating to two segments
	parts := strings.SplitN(path, "/", 4)
	if len(parts) > 3 {
		return strings.Join(parts[:3], "/") + "/..."
	}
	return path
}

func CollectRequestStats() func(http.Handler) http.Handler {
	generator, _ := aca.NewDefaultGenerator()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestId := generator.Generate()
			startRequest(requestId)

			writer := &StatsResponseWriter{
				ResponseWriter: w,
				requestID:      requestId,
			}

			next.ServeHTTP(writer, r.WithContext(context.WithValue(r.Context(), requestIdKey, requestId)))

			duration, queries := func() (time.Duration, int32) {
				inFlightRequestsMu.RLock()
				defer inFlightRequestsMu.RUnlock()
				details, ok := inFlightRequests[requestId]
				if !ok {
					return 0, 0
				}
				return time.Since(details.start), details.queries.Load()
			}()

			path := normalizePath(r.URL.Path)
			status := writer.statusCode()
			httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
			dbQueriesPerRequest.WithLabelValues(path).Observe(float64(queries))

			go recordRequestMetric(r.URL.Path, requestId, duration, queries)

			writer.Finish(duration, queries)
			pruneRequest(requestId)
		})
	}
}

func startRequest(requestId string) {
	inFlightRequestsMu.Lock()
	defer inFlightRequestsMu.Unlock()
	inFlightRequests[requestId] = &request{
		start: time.Now(),
	}
}

func pruneRequest(requestId string) {
	inFlightRequestsMu.Lock()
	defer inFlightRequestsMu.Unlock()
	delete(inFlightRequests, requestId)
}

func LogQuery(ctx context.Context) {
	requestId, ok := ctx.Value(requestIdKey).(string)
	if !ok {
		return
	}

	inFlightRequestsMu.RLock()
	details, ok := inFlightRequests[requestId]
	inFlightRequestsMu.RUnlock()
	if !ok {
		return
	}

	details.queries.Add(1)
}

var statsPlaceholder = []byte("[[STATS_GO_HERE]]")

type StatsResponseWriter struct {
	http.ResponseWriter
	buffer      []byte
	passthrough bool
	wroteHeader bool
	code        int
	requestID   string
}

func (w *StatsResponseWriter) Write(b []byte) (int, error) {
	// Only HTML pages can contain the stats placeholder, so only they need
	// buffering; everything else (assets, media, feeds) is written straight
	// through. The decision is made on the first non-empty write.
	if !w.passthrough && len(w.buffer) == 0 && len(b) > 0 {
		w.passthrough = !isHTML(w.Header().Get("Content-Type"), b)
	}

	if w.passthrough {
		return w.ResponseWriter.Write(b)
	}

	w.buffer = append(w.buffer, b...)
	return len(b), nil
}

func isHTML(contentType string, firstChunk []byte) bool {
	if contentType == "" {
		contentType = http.DetectContentType(firstChunk)
	}
	return strings.HasPrefix(contentType, "text/html")
}

func (w *StatsResponseWriter) WriteHeader(statusCode int) {
	w.wroteHeader = true
	w.code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *StatsResponseWriter) statusCode() string {
	if w.code == 0 {
		return "200"
	}
	return strconv.Itoa(w.code)
}

// Flush implements http.Flusher so that non-buffered responses can stream.
// For buffered pages only the headers can be flushed early; the body follows
// when the request completes.
func (w *StatsResponseWriter) Flush() {
	w.wroteHeader = true
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Finish completes the response: buffered HTML pages have the request stats
// injected before being written out; anything else is already done.
func (w *StatsResponseWriter) Finish(duration time.Duration, queries int32) {
	if w.passthrough || len(w.buffer) == 0 {
		return
	}

	w.buffer = bytes.Replace(w.buffer, statsPlaceholder, statsHTML(w.requestID, duration, queries), 1)

	if !w.wroteHeader {
		w.ResponseWriter.WriteHeader(http.StatusOK)
	}
	_, _ = w.ResponseWriter.Write(w.buffer)
}

func statsHTML(requestID string, duration time.Duration, queries int32) []byte {
	if duration == 0 {
		return []byte("There would be request stats here, but I seem to have misplaced them...")
	}

	shortCommit := buildVersion
	if buildVersion == "" {
		shortCommit = "unknown"
	} else if len(buildVersion) > 7 {
		shortCommit = buildVersion[:7]
	}

	p := message.NewPrinter(language.English)
	return []byte(p.Sprintf(
		`Request ID <code>%s</code> served by chameth.com <code>%s</code> in %dμs using %d db queries`,
		requestID,
		shortCommit,
		duration.Microseconds(),
		queries,
	))
}
