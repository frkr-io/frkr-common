// Package metrics provides Prometheus metrics for frkr gateways.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Common metrics shared by all gateways
var (
	// Up indicates if the service is healthy (1 = healthy, 0 = unhealthy)
	Up = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "frkr",
		Name:      "up",
		Help:      "1 if the service is healthy, 0 otherwise",
	})

	// Info provides service metadata as labels
	Info = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "frkr",
		Name:      "info",
		Help:      "Service metadata",
	}, []string{"service", "version"})

	// AuthFailuresTotal counts authentication failures
	AuthFailuresTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "frkr",
		Name:      "auth_failures_total",
		Help:      "Total number of authentication failures",
	}, []string{"service", "reason"})
)

// Registry is the default Prometheus registry for frkr metrics
var Registry = prometheus.NewRegistry()

func init() {
	// Register common metrics
	Registry.MustRegister(Up)
	Registry.MustRegister(Info)
	Registry.MustRegister(AuthFailuresTotal)

	// Register Go runtime metrics (goroutines, memory, GC)
	Registry.MustRegister(collectors.NewGoCollector())

	// Register process metrics (CPU, open FDs, etc.)
	Registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}

// Handler returns an HTTP handler for the /metrics endpoint
func Handler() http.Handler {
	return promhttp.HandlerFor(Registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// Register adds a collector to the frkr metrics registry
func Register(c prometheus.Collector) error {
	return Registry.Register(c)
}

// MustRegister adds collectors to the frkr metrics registry, panicking on error
func MustRegister(cs ...prometheus.Collector) {
	Registry.MustRegister(cs...)
}

// SetServiceInfo sets the service info metric
func SetServiceInfo(service, version string) {
	Info.WithLabelValues(service, version).Set(1)
}

// SetHealthy sets the up metric to indicate healthy status
func SetHealthy() {
	Up.Set(1)
}

// SetUnhealthy sets the up metric to indicate unhealthy status
func SetUnhealthy() {
	Up.Set(0)
}

// RecordAuthFailure increments the auth failure counter
func RecordAuthFailure(service, reason string) {
	AuthFailuresTotal.WithLabelValues(service, reason).Inc()
}
