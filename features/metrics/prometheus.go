package metrics

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsPort = flag.Int("metrics-port", 9090, "Port to serve Prometheus metrics on (0 to disable)")
)

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
