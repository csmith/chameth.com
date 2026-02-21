package metrics

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsPort = flag.Int("metrics-port", 9090, "Port to serve Prometheus metrics on (0 to disable)")

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	dbQueriesPerRequest = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_db_queries",
			Help:    "Number of database queries per HTTP request.",
			Buckets: []float64{0, 1, 2, 3, 5, 8, 13, 21, 34, 55},
		},
		[]string{"path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(dbQueriesPerRequest)
}

func StartMetricsServer() {
	if *metricsPort == 0 {
		slog.Info("Metrics server disabled")
		return
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *metricsPort),
		Handler: mux,
	}

	go func() {
		slog.Info(fmt.Sprintf("Metrics server listening on http://0.0.0.0:%d/metrics", *metricsPort))
		if err := server.ListenAndServe(); err != nil {
			slog.Error("Metrics server failed", "error", err)
		}
	}()
}
