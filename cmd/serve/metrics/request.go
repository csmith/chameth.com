package metrics

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"runtime/debug"

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

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		buildVersion = "master"
		return
	}

	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			buildVersion = setting.Value
			return
		}
	}

	buildVersion = "master"
}

type request struct {
	start   time.Time
	queries atomic.Int32
}

func CollectRequestStats() func(http.Handler) http.Handler {
	generator, _ := aca.NewDefaultGenerator()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestId := generator.Generate()
			startRequest(requestId)
			defer pruneRequest(requestId)

			writer := &StatsResponseWriter{
				ResponseWriter: w,
				requestID:      requestId,
			}

			next.ServeHTTP(writer, r.WithContext(context.WithValue(r.Context(), requestIdKey, requestId)))

			slog.Info("Finished request", "rid", requestId)
			writer.Flush()
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
	requestID   string
}

func (w *StatsResponseWriter) Write(b []byte) (int, error) {
	w.buffer = append(w.buffer, b...)
	return len(b), nil
}

func (w *StatsResponseWriter) WriteHeader(statusCode int) {
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *StatsResponseWriter) Flush() {
	content := injectStats(w.requestID, string(w.buffer))
	if !w.wroteHeader {
		w.ResponseWriter.WriteHeader(http.StatusOK)
	}
	w.ResponseWriter.Write([]byte(content))
}

func injectStats(requestID string, content string) string {
	duration, queries := func() (time.Duration, int32) {
		inFlightRequestsMu.RLock()
		defer inFlightRequestsMu.RUnlock()
		details, ok := inFlightRequests[requestID]
		if !ok {
			return 0, 0
		}

		return time.Since(details.start), details.queries.Load()
	}()

	if duration == 0 {
		return strings.Replace(content, "[[STATS_GO_HERE]]", "There would be request stats here, but I seem to have misplaced them...", 1)
	}

	shortCommit := buildVersion
	if len(buildVersion) > 7 {
		shortCommit = buildVersion[:7]
	}

	p := message.NewPrinter(language.English)
	return strings.Replace(
		content,
		"[[STATS_GO_HERE]]",
		p.Sprintf(
			`Request ID <code>%s</code> served by chameth.com build <code>%s</code> in %dμs using %d db queries`,
			requestID,
			shortCommit,
			duration.Microseconds(),
			queries,
		),
		1,
	)
}
