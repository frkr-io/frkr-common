package metrics_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/frkr-io/frkr-common/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsHandler_Smoke(t *testing.T) {
	// Register ingest metrics to have something to scrape
	metrics.RegisterIngestMetrics()
	metrics.SetServiceInfo("test-service", "1.0.0")
	metrics.SetHealthy()

	// Create a test server with the metrics handler
	handler := metrics.Handler()
	server := httptest.NewServer(handler)
	defer server.Close()

	// Make a request to the metrics endpoint
	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyStr := string(body)

	// Verify common metrics are present
	assert.Contains(t, bodyStr, "frkr_up", "should contain frkr_up metric")
	assert.Contains(t, bodyStr, "frkr_info", "should contain frkr_info metric")

	// Verify Go runtime metrics are present
	assert.Contains(t, bodyStr, "go_goroutines", "should contain Go runtime metrics")

	// Verify process metrics are present
	assert.Contains(t, bodyStr, "process_", "should contain process metrics")
}

func TestMetricsHandler_IngestMetrics(t *testing.T) {
	// Record some ingest activity
	metrics.RecordIngestRequest("POST", "/ingest", "200", 0.05)
	metrics.RecordMessagePublished("test-stream")
	metrics.RecordPublishError("test-stream", "timeout")

	// Create test server
	handler := metrics.Handler()
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyStr := string(body)

	// Verify ingest-specific metrics
	assert.Contains(t, bodyStr, "frkr_ingest_requests_total", "should contain ingest requests counter")
	assert.Contains(t, bodyStr, "frkr_ingest_request_duration_seconds", "should contain request duration histogram")
	assert.Contains(t, bodyStr, "frkr_ingest_messages_published_total", "should contain messages published counter")
	assert.Contains(t, bodyStr, "frkr_ingest_publish_errors_total", "should contain publish errors counter")
}

func TestMetricsHandler_StreamingMetrics(t *testing.T) {
	// Register streaming metrics
	metrics.RegisterStreamingMetrics()

	// Record some streaming activity
	metrics.RecordStreamOpened()
	metrics.RecordMessageDelivered("test-stream")
	metrics.RecordStreamDuration("test-stream", 30.5)
	metrics.RecordStreamClosed()

	// Create test server
	handler := metrics.Handler()
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyStr := string(body)

	// Verify streaming-specific metrics
	assert.Contains(t, bodyStr, "frkr_streaming_active_streams", "should contain active streams gauge")
	assert.Contains(t, bodyStr, "frkr_streaming_messages_delivered_total", "should contain messages delivered counter")
	assert.Contains(t, bodyStr, "frkr_streaming_stream_duration_seconds", "should contain stream duration histogram")
}

func TestMetricsHandler_AuthFailures(t *testing.T) {
	// Record auth failures
	metrics.RecordAuthFailure("test-gateway", "invalid_token")
	metrics.RecordAuthFailure("test-gateway", "expired_token")

	// Create test server
	handler := metrics.Handler()
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyStr := string(body)

	// Verify auth failure counter
	assert.Contains(t, bodyStr, "frkr_auth_failures_total", "should contain auth failures counter")
	assert.True(t, strings.Contains(bodyStr, "invalid_token") || strings.Contains(bodyStr, "expired_token"),
		"should contain failure reasons in labels")
}

func TestHealthStatus(t *testing.T) {
	// Test healthy status
	metrics.SetHealthy()

	handler := metrics.Handler()
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// frkr_up should be 1
	assert.Contains(t, string(body), "frkr_up 1", "frkr_up should be 1 when healthy")

	// Test unhealthy status
	metrics.SetUnhealthy()

	resp2, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp2.Body.Close()

	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)

	// frkr_up should be 0
	assert.Contains(t, string(body2), "frkr_up 0", "frkr_up should be 0 when unhealthy")
}
