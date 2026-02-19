package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/hannan/wirefight/go-service/internal/logic"
	"github.com/hannan/wirefight/go-service/internal/metrics"
)

// Handler handles REST API requests
type Handler struct {
	metrics *metrics.Collector
	debug   bool
}

// NewHandler creates a new REST handler
func NewHandler(m *metrics.Collector, debug bool) *Handler {
	return &Handler{
		metrics: m,
		debug:   debug,
	}
}

// ComputeRequest represents the REST API request format
type ComputeRequest struct {
	RequestID        string `json:"request_id"`
	Mode             string `json:"mode"`
	WorkFactor       int    `json:"work_factor"`
	PayloadSizeBytes int    `json:"payload_size_bytes"`
}

// ComputeResponse represents the REST API response format
type ComputeResponse struct {
	RequestID          string  `json:"request_id"`
	Mode               string  `json:"mode"`
	WorkFactor         int     `json:"work_factor"`
	PayloadSizeBytes   int     `json:"payload_size_bytes"`
	Result             string  `json:"result"`
	ServerProcessingMs float64 `json:"server_processing_ms"`
	Protocol           string  `json:"protocol"`
	Timestamp          string  `json:"timestamp"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// HandleCompute processes REST compute requests
func (h *Handler) HandleCompute(w http.ResponseWriter, r *http.Request) {
	h.metrics.IncrementRESTRequests()

	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		h.metrics.IncrementRESTErrors()
		return
	}

	var req ComputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if h.debug {
			log.Printf("[REST] Failed to decode request: %v", err)
		}
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		h.metrics.IncrementRESTErrors()
		return
	}

	// Convert to internal logic request
	logicReq := logic.Request{
		RequestID:        req.RequestID,
		Mode:             logic.Mode(req.Mode),
		WorkFactor:       req.WorkFactor,
		PayloadSizeBytes: req.PayloadSizeBytes,
	}

	// Execute shared logic
	logicResp, err := logic.Execute(logicReq, "rest")
	if err != nil {
		if h.debug {
			log.Printf("[REST] Logic execution failed: %v", err)
		}
		h.sendError(w, http.StatusBadRequest, err.Error())
		h.metrics.IncrementRESTErrors()
		return
	}

	// Convert to REST response
	resp := ComputeResponse{
		RequestID:          logicResp.RequestID,
		Mode:               string(logicResp.Mode),
		WorkFactor:         logicResp.WorkFactor,
		PayloadSizeBytes:   logicResp.PayloadSizeBytes,
		Result:             logicResp.Result,
		ServerProcessingMs: logicResp.ServerProcessingMs,
		Protocol:           logicResp.Protocol,
		Timestamp:          logicResp.Timestamp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// sendError sends a JSON error response
func (h *Handler) sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}
