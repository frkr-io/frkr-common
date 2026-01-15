// Package metrics provides Prometheus metrics for frkr gateways.
// This file contains metrics specific to the Ingest Gateway.
package metrics

import "github.com/prometheus/client_golang/prometheus"

// Ingest Gateway specific metrics
var (
	// IngestRequestsTotal counts total HTTP requests received
	IngestRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "frkr",
		Subsystem: "ingest",
		Name:      "requests_total",
		Help:      "Total number of HTTP requests received",
	}, []string{"method", "path", "status"})

	// IngestRequestDuration measures request latency
	IngestRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "frkr",
		Subsystem: "ingest",
		Name:      "request_duration_seconds",
		Help:      "HTTP request latency in seconds",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path"})

	// IngestMessagesPublishedTotal counts messages published to Kafka
	IngestMessagesPublishedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "frkr",
		Subsystem: "ingest",
		Name:      "messages_published_total",
		Help:      "Total number of messages published to Kafka",
	}, []string{"stream_id"})

	// IngestPublishErrorsTotal counts Kafka publish failures
	IngestPublishErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "frkr",
		Subsystem: "ingest",
		Name:      "publish_errors_total",
		Help:      "Total number of Kafka publish failures",
	}, []string{"stream_id", "error_type"})
)

// RegisterIngestMetrics registers all ingest gateway metrics
func RegisterIngestMetrics() {
	MustRegister(
		IngestRequestsTotal,
		IngestRequestDuration,
		IngestMessagesPublishedTotal,
		IngestPublishErrorsTotal,
	)
}

// RecordIngestRequest records an HTTP request with its outcome
func RecordIngestRequest(method, path, status string, durationSeconds float64) {
	IngestRequestsTotal.WithLabelValues(method, path, status).Inc()
	IngestRequestDuration.WithLabelValues(method, path).Observe(durationSeconds)
}

// RecordMessagePublished records a successful message publish
func RecordMessagePublished(streamID string) {
	IngestMessagesPublishedTotal.WithLabelValues(streamID).Inc()
}

// RecordPublishError records a Kafka publish failure
func RecordPublishError(streamID, errorType string) {
	IngestPublishErrorsTotal.WithLabelValues(streamID, errorType).Inc()
}
