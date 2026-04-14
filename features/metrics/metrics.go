package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
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

	contactSubmissionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "contact_submissions_total",
			Help: "Total contact form submissions. Each cause label is true/false.",
		},
		append([]string{"method"}, contactCauseLabelNames()...),
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(dbQueriesPerRequest)
	prometheus.MustRegister(contactSubmissionsTotal)
}
