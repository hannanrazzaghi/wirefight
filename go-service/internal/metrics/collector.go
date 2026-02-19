package metrics

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// Collector tracks basic metrics
type Collector struct {
	startTime time.Time

	// Per-protocol counters
	restRequests    atomic.Int64
	jsonrpcRequests atomic.Int64
	grpcRequests    atomic.Int64

	// Error counters
	restErrors    atomic.Int64
	jsonrpcErrors atomic.Int64
	grpcErrors    atomic.Int64
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		startTime: time.Now(),
	}
}

// IncrementRESTRequests increments REST request counter
func (c *Collector) IncrementRESTRequests() {
	c.restRequests.Add(1)
}

// IncrementRESTErrors increments REST error counter
func (c *Collector) IncrementRESTErrors() {
	c.restErrors.Add(1)
}

// IncrementJSONRPCRequests increments JSON-RPC request counter
func (c *Collector) IncrementJSONRPCRequests() {
	c.jsonrpcRequests.Add(1)
}

// IncrementJSONRPCErrors increments JSON-RPC error counter
func (c *Collector) IncrementJSONRPCErrors() {
	c.jsonrpcErrors.Add(1)
}

// IncrementGRPCRequests increments gRPC request counter
func (c *Collector) IncrementGRPCRequests() {
	c.grpcRequests.Add(1)
}

// IncrementGRPCErrors increments gRPC error counter
func (c *Collector) IncrementGRPCErrors() {
	c.grpcErrors.Add(1)
}

// Snapshot represents metrics at a point in time
type Snapshot struct {
	UptimeSeconds   float64 `json:"uptime_seconds"`
	RESTRequests    int64   `json:"rest_requests"`
	RESTErrors      int64   `json:"rest_errors"`
	JSONRPCRequests int64   `json:"jsonrpc_requests"`
	JSONRPCErrors   int64   `json:"jsonrpc_errors"`
	GRPCRequests    int64   `json:"grpc_requests"`
	GRPCErrors      int64   `json:"grpc_errors"`
	TotalRequests   int64   `json:"total_requests"`
	TotalErrors     int64   `json:"total_errors"`
}

// Snapshot returns current metrics snapshot
func (c *Collector) Snapshot() Snapshot {
	rest := c.restRequests.Load()
	jsonrpc := c.jsonrpcRequests.Load()
	grpc := c.grpcRequests.Load()

	restErr := c.restErrors.Load()
	jsonrpcErr := c.jsonrpcErrors.Load()
	grpcErr := c.grpcErrors.Load()

	return Snapshot{
		UptimeSeconds:   time.Since(c.startTime).Seconds(),
		RESTRequests:    rest,
		RESTErrors:      restErr,
		JSONRPCRequests: jsonrpc,
		JSONRPCErrors:   jsonrpcErr,
		GRPCRequests:    grpc,
		GRPCErrors:      grpcErr,
		TotalRequests:   rest + jsonrpc + grpc,
		TotalErrors:     restErr + jsonrpcErr + grpcErr,
	}
}

// HTTPHandler returns an HTTP handler for metrics endpoint
func (c *Collector) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot := c.Snapshot()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snapshot)
	}
}

// HealthHandler returns a simple health check handler
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	}
}
