package logic

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// Mode defines the type of workload
type Mode string

const (
	ModeCPU Mode = "cpu"
	ModeIO  Mode = "io"
)

// Request represents the semantic request structure
type Request struct {
	RequestID        string `json:"request_id"`
	Mode             Mode   `json:"mode"`
	WorkFactor       int    `json:"work_factor"`
	PayloadSizeBytes int    `json:"payload_size_bytes"`
}

// Response represents the semantic response structure
type Response struct {
	RequestID            string  `json:"request_id"`
	Mode                 Mode    `json:"mode"`
	WorkFactor           int     `json:"work_factor"`
	PayloadSizeBytes     int     `json:"payload_size_bytes"`
	Result               string  `json:"result"`
	ServerProcessingMs   float64 `json:"server_processing_ms"`
	Protocol             string  `json:"protocol"`
	Timestamp            string  `json:"timestamp"`
}

// Execute is the shared business logic called by all protocols
// It measures only the time spent in this function (pure processing time)
func Execute(req Request, protocol string) (*Response, error) {
	start := time.Now()

	// Validate request
	if req.RequestID == "" {
		return nil, fmt.Errorf("request_id is required")
	}
	if req.Mode != ModeCPU && req.Mode != ModeIO {
		return nil, fmt.Errorf("invalid mode: %s (must be 'cpu' or 'io')", req.Mode)
	}
	if req.WorkFactor < 0 {
		return nil, fmt.Errorf("work_factor must be non-negative")
	}
	if req.PayloadSizeBytes < 0 {
		return nil, fmt.Errorf("payload_size_bytes must be non-negative")
	}

	// Execute workload
	var result string
	switch req.Mode {
	case ModeCPU:
		result = executeCPUWork(req.WorkFactor, req.PayloadSizeBytes)
	case ModeIO:
		result = executeIOWork(req.WorkFactor, req.PayloadSizeBytes)
	default:
		return nil, fmt.Errorf("unsupported mode: %s", req.Mode)
	}

	elapsed := time.Since(start)

	return &Response{
		RequestID:            req.RequestID,
		Mode:                 req.Mode,
		WorkFactor:           req.WorkFactor,
		PayloadSizeBytes:     req.PayloadSizeBytes,
		Result:               result,
		ServerProcessingMs:   float64(elapsed.Microseconds()) / 1000.0,
		Protocol:             protocol,
		Timestamp:            time.Now().UTC().Format(time.RFC3339Nano),
	}, nil
}

// executeCPUWork performs deterministic CPU-bound work
func executeCPUWork(iterations int, payloadSize int) string {
	data := make([]byte, 32)
	copy(data, []byte("benchmark-seed-data"))

	// Perform SHA-256 hashing in a loop
	for i := 0; i < iterations; i++ {
		hash := sha256.Sum256(data)
		data = hash[:]

		// Optionally add JSON marshal/unmarshal overhead
		if payloadSize > 0 && i%10 == 0 {
			payload := generatePayload(payloadSize)
			jsonData, _ := json.Marshal(payload)
			var dummy map[string]interface{}
			json.Unmarshal(jsonData, &dummy)
		}
	}

	return fmt.Sprintf("%x", data[:8])
}

// executeIOWork simulates I/O-bound work
func executeIOWork(sleepMs int, payloadSize int) string {
	// Break sleep into smaller chunks for more realistic behavior
	chunks := 10
	if sleepMs < 10 {
		chunks = 1
	}
	chunkDuration := time.Duration(sleepMs/chunks) * time.Millisecond

	for i := 0; i < chunks; i++ {
		time.Sleep(chunkDuration)

		// Optionally add payload processing
		if payloadSize > 0 && i%3 == 0 {
			payload := generatePayload(payloadSize)
			jsonData, _ := json.Marshal(payload)
			var dummy map[string]interface{}
			json.Unmarshal(jsonData, &dummy)
		}
	}

	return fmt.Sprintf("slept_%dms", sleepMs)
}

// generatePayload creates a dummy payload of approximately the specified size
func generatePayload(sizeBytes int) map[string]interface{} {
	// Create a payload with repeated data to reach target size
	// Each character is roughly 1 byte in JSON
	dataSize := sizeBytes / 2 // Rough approximation for JSON overhead
	if dataSize < 1 {
		dataSize = 1
	}

	data := make([]byte, dataSize)
	for i := range data {
		data[i] = byte('A' + (i % 26))
	}

	return map[string]interface{}{
		"data":      string(data),
		"size":      sizeBytes,
		"timestamp": time.Now().Unix(),
	}
}
