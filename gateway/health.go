// Package gateway provides common types and utilities for frkr gateway services.
package gateway

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// HealthResponse is the structured health check response (RFC Health Check format)
type HealthResponse struct {
	Status  string           `json:"status"` // "healthy", "degraded", "unhealthy"
	Checks  map[string]Check `json:"checks,omitempty"`
	Version string           `json:"version,omitempty"`
	Uptime  string           `json:"uptime,omitempty"`
}

// Check represents a single health check result
type Check struct {
	Status  string `json:"status"` // "pass", "fail"
	Message string `json:"message,omitempty"`
}

// StatusResponse provides detailed gateway status
type StatusResponse struct {
	Service  string `json:"service"`
	Version  string `json:"version"`
	Port     int    `json:"port"`
	Uptime   string `json:"uptime"`
	Database string `json:"database"` // Connection info (sanitized)
	Broker   string `json:"broker"`
	Ready    bool   `json:"ready"`
}

// HealthChecker manages health state for a gateway service
type HealthChecker struct {
	mu            sync.RWMutex
	dbHealthy     bool
	brokerHealthy bool
	startTime     time.Time
	version       string
	serviceName   string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(serviceName, version string) *HealthChecker {
	return &HealthChecker{
		startTime:   time.Now(),
		version:     version,
		serviceName: serviceName,
	}
}

// StartHealthCheckLoop starts a background goroutine that periodically checks dependencies
func (hc *HealthChecker) StartHealthCheckLoop(db *sql.DB, brokerURL string) {
	// Initial check
	hc.CheckDependencies(db, brokerURL)

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			hc.CheckDependencies(db, brokerURL)
		}
	}()
}

// CheckDependencies performs health checks on database and broker
func (hc *HealthChecker) CheckDependencies(db *sql.DB, brokerURL string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	// Check database
	if db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		hc.dbHealthy = db.PingContext(ctx) == nil
		cancel()
	} else {
		hc.dbHealthy = false
	}

	// Check broker
	if brokerURL != "" {
		conn, err := kafka.Dial("tcp", brokerURL)
		if err == nil {
			conn.Close()
			hc.brokerHealthy = true
		} else {
			hc.brokerHealthy = false
		}
	} else {
		hc.brokerHealthy = false
	}
}

// IsReady returns true if all dependencies are healthy
func (hc *HealthChecker) IsReady() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.dbHealthy && hc.brokerHealthy
}

// GetHealthState returns the current health state
func (hc *HealthChecker) GetHealthState() (dbHealthy, brokerHealthy bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.dbHealthy, hc.brokerHealthy
}

// HandleHealth is the default health endpoint (alias for readiness)
func (hc *HealthChecker) HandleHealth(w http.ResponseWriter, r *http.Request) {
	hc.HandleReadiness(w, r)
}

// HandleReadiness checks if the gateway is ready to serve traffic
// Returns 503 if dependencies are unhealthy
func (hc *HealthChecker) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	dbHealthy, brokerHealthy := hc.GetHealthState()

	dbCheck := Check{Status: boolToStatus(dbHealthy)}
	brokerCheck := Check{Status: boolToStatus(brokerHealthy)}

	if !dbHealthy {
		dbCheck.Message = "database connection failed"
	}
	if !brokerHealthy {
		brokerCheck.Message = "broker connection failed"
	}

	checks := map[string]Check{
		"database": dbCheck,
		"broker":   brokerCheck,
	}

	status := "healthy"
	httpStatus := http.StatusOK

	if !dbHealthy || !brokerHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	resp := HealthResponse{
		Status:  status,
		Checks:  checks,
		Version: hc.version,
		Uptime:  time.Since(hc.startTime).Round(time.Second).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(resp)
}

// HandleLiveness checks if the gateway process is alive
// Always returns 200 if the process is running
func (hc *HealthChecker) HandleLiveness(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:  "healthy",
		Version: hc.version,
		Uptime:  time.Since(hc.startTime).Round(time.Second).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleStatus returns detailed service status
func (hc *HealthChecker) HandleStatus(w http.ResponseWriter, r *http.Request, port int, dbURL, brokerURL string) {
	resp := StatusResponse{
		Service:  hc.serviceName,
		Version:  hc.version,
		Port:     port,
		Uptime:   time.Since(hc.startTime).Round(time.Second).String(),
		Database: SanitizeURL(dbURL),
		Broker:   brokerURL,
		Ready:    hc.IsReady(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RegisterHealthEndpoints registers standard health endpoints on a ServeMux
func (hc *HealthChecker) RegisterHealthEndpoints(mux *http.ServeMux, port int, dbURL, brokerURL string) {
	mux.HandleFunc("/health", hc.HandleHealth)
	mux.HandleFunc("/health/ready", hc.HandleReadiness)
	mux.HandleFunc("/health/live", hc.HandleLiveness)
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		hc.HandleStatus(w, r, port, dbURL, brokerURL)
	})
}

// Helper functions

func boolToStatus(b bool) string {
	if b {
		return "pass"
	}
	return "fail"
}

// SanitizeURL removes password from URL for safe display/logging
func SanitizeURL(url string) string {
	if strings.Contains(url, "@") {
		parts := strings.SplitN(url, "@", 2)
		if len(parts) == 2 {
			protoIdx := strings.Index(parts[0], "://")
			if protoIdx != -1 {
				proto := parts[0][:protoIdx+3]
				userPart := parts[0][protoIdx+3:]
				if colonIdx := strings.Index(userPart, ":"); colonIdx != -1 {
					userPart = userPart[:colonIdx]
				}
				return proto + userPart + "@" + parts[1]
			}
		}
	}
	return url
}
