// Package metrics provides Prometheus metrics for frkr gateways.
// This file contains metrics specific to the Streaming Gateway.
package metrics

import "github.com/prometheus/client_golang/prometheus"

// Streaming Gateway specific metrics
var (
	// StreamingActiveStreams tracks currently open gRPC streams
	StreamingActiveStreams = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "frkr",
		Subsystem: "streaming",
		Name:      "active_streams",
		Help:      "Number of currently open gRPC streams",
	})

	// StreamingMessagesDeliveredTotal counts messages sent to clients
	StreamingMessagesDeliveredTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "frkr",
		Subsystem: "streaming",
		Name:      "messages_delivered_total",
		Help:      "Total number of messages delivered to CLI clients",
	}, []string{"stream_id"})

	// StreamingStreamDuration measures stream session duration
	StreamingStreamDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "frkr",
		Subsystem: "streaming",
		Name:      "stream_duration_seconds",
		Help:      "Duration of stream sessions in seconds",
		Buckets:   []float64{1, 5, 15, 30, 60, 120, 300, 600, 1800, 3600},
	}, []string{"stream_id"})

	// StreamingConsumerLag tracks Kafka consumer lag (if available)
	StreamingConsumerLag = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "frkr",
		Subsystem: "streaming",
		Name:      "consumer_lag",
		Help:      "Kafka consumer lag (messages behind)",
	}, []string{"stream_id", "partition"})
)

// RegisterStreamingMetrics registers all streaming gateway metrics
func RegisterStreamingMetrics() {
	MustRegister(
		StreamingActiveStreams,
		StreamingMessagesDeliveredTotal,
		StreamingStreamDuration,
		StreamingConsumerLag,
	)
}

// RecordStreamOpened increments active stream count
func RecordStreamOpened() {
	StreamingActiveStreams.Inc()
}

// RecordStreamClosed decrements active stream count
func RecordStreamClosed() {
	StreamingActiveStreams.Dec()
}

// RecordMessageDelivered records a message delivered to a client
func RecordMessageDelivered(streamID string) {
	StreamingMessagesDeliveredTotal.WithLabelValues(streamID).Inc()
}

// RecordStreamDuration records the duration of a stream session
func RecordStreamDuration(streamID string, durationSeconds float64) {
	StreamingStreamDuration.WithLabelValues(streamID).Observe(durationSeconds)
}

// SetConsumerLag sets the consumer lag for a stream/partition
func SetConsumerLag(streamID, partition string, lag float64) {
	StreamingConsumerLag.WithLabelValues(streamID, partition).Set(lag)
}
