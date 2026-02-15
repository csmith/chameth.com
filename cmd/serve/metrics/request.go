package metrics

import (
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

var (
	requestIdKey = struct{}{}

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

			writer.Flush(duration, queries)
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
	defer inFlightRequestsMu.RUnlock()
	details, ok := inFlightRequests[requestId]
	if !ok {
		return
	}

	details.queries.Add(1)
}

type StatsResponseWriter struct {
	http.ResponseWriter
	buffer      []byte
	wroteHeader bool
	code        int
	requestID   string
}

func (w *StatsResponseWriter) Write(b []byte) (int, error) {
	w.buffer = append(w.buffer, b...)
	return len(b), nil
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

func (w *StatsResponseWriter) Flush(duration time.Duration, queries int32) {
	content := injectStats(w.requestID, string(w.buffer), duration, queries)
	if !w.wroteHeader {
		w.ResponseWriter.WriteHeader(http.StatusOK)
	}
	w.ResponseWriter.Write([]byte(content))
}

func injectStats(requestID string, content string, duration time.Duration, queries int32) string {
	if duration == 0 {
		return strings.Replace(content, "[[STATS_GO_HERE]]", "There would be request stats here, but I seem to have misplaced them...", 1)
	}

	shortCommit := buildVersion
	if buildVersion == "" {
		shortCommit = "unknown"
	} else if len(buildVersion) > 7 {
		shortCommit = buildVersion[:7]
	}

	p := message.NewPrinter(language.English)
	return strings.Replace(
		content,
		"[[STATS_GO_HERE]]",
		p.Sprintf(
			`Request ID <code>%s</code> served by chameth.com <code>%s</code> in %dμs using %d db queries`,
			requestID,
			shortCommit,
			duration.Microseconds(),
			queries,
		),
		1,
	)
}
