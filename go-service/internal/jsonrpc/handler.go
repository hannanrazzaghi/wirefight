package jsonrpc

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hannanrazzaghi/wirefight/go-service/internal/logic"
	"github.com/hannanrazzaghi/wirefight/go-service/internal/metrics"
)

// Handler handles JSON-RPC 2.0 requests
type Handler struct {
	metrics *metrics.Collector
	debug   bool
}

// NewHandler creates a new JSON-RPC handler
func NewHandler(m *metrics.Collector, debug bool) *Handler {
	return &Handler{
		metrics: m,
		debug:   debug,
	}
}

// Request represents a JSON-RPC 2.0 request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

// Response represents a JSON-RPC 2.0 response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// Error represents a JSON-RPC 2.0 error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Standard JSON-RPC error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// ComputeParams represents the parameters for the compute method
type ComputeParams struct {
	RequestID        string `json:"request_id"`
	Mode             string `json:"mode"`
	WorkFactor       int    `json:"work_factor"`
	PayloadSizeBytes int    `json:"payload_size_bytes"`
}

// ComputeResult represents the result of the compute method
type ComputeResult struct {
	RequestID          string  `json:"request_id"`
	Mode               string  `json:"mode"`
	WorkFactor         int     `json:"work_factor"`
	PayloadSizeBytes   int     `json:"payload_size_bytes"`
	Result             string  `json:"result"`
	ServerProcessingMs float64 `json:"server_processing_ms"`
	Protocol           string  `json:"protocol"`
	Timestamp          string  `json:"timestamp"`
}

// HandleRPC processes JSON-RPC 2.0 requests
func (h *Handler) HandleRPC(w http.ResponseWriter, r *http.Request) {
	h.metrics.IncrementJSONRPCRequests()

	if r.Method != http.MethodPost {
		h.sendError(w, nil, MethodNotFound, "method not allowed")
		h.metrics.IncrementJSONRPCErrors()
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if h.debug {
			log.Printf("[JSON-RPC] Failed to decode request: %v", err)
		}
		h.sendError(w, nil, ParseError, "parse error")
		h.metrics.IncrementJSONRPCErrors()
		return
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		h.sendError(w, req.ID, InvalidRequest, "invalid jsonrpc version")
		h.metrics.IncrementJSONRPCErrors()
		return
	}

	// Route to method
	if req.Method != "compute" {
		h.sendError(w, req.ID, MethodNotFound, fmt.Sprintf("method not found: %s", req.Method))
		h.metrics.IncrementJSONRPCErrors()
		return
	}

	// Handle compute method
	h.handleCompute(w, req)
}

func (h *Handler) handleCompute(w http.ResponseWriter, req Request) {
	var params ComputeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		if h.debug {
			log.Printf("[JSON-RPC] Failed to unmarshal params: %v", err)
		}
		h.sendError(w, req.ID, InvalidParams, "invalid params")
		h.metrics.IncrementJSONRPCErrors()
		return
	}

	// Convert to internal logic request
	logicReq := logic.Request{
		RequestID:        params.RequestID,
		Mode:             logic.Mode(params.Mode),
		WorkFactor:       params.WorkFactor,
		PayloadSizeBytes: params.PayloadSizeBytes,
	}

	// Execute shared logic
	logicResp, err := logic.Execute(logicReq, "jsonrpc")
	if err != nil {
		if h.debug {
			log.Printf("[JSON-RPC] Logic execution failed: %v", err)
		}
		h.sendError(w, req.ID, InvalidParams, err.Error())
		h.metrics.IncrementJSONRPCErrors()
		return
	}

	// Convert to JSON-RPC result
	result := ComputeResult{
		RequestID:          logicResp.RequestID,
		Mode:               string(logicResp.Mode),
		WorkFactor:         logicResp.WorkFactor,
		PayloadSizeBytes:   logicResp.PayloadSizeBytes,
		Result:             logicResp.Result,
		ServerProcessingMs: logicResp.ServerProcessingMs,
		Protocol:           logicResp.Protocol,
		Timestamp:          logicResp.Timestamp,
	}

	h.sendSuccess(w, req.ID, result)
}

func (h *Handler) sendSuccess(w http.ResponseWriter, id interface{}, result interface{}) {
	resp := Response{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) sendError(w http.ResponseWriter, id interface{}, code int, message string) {
	resp := Response{
		JSONRPC: "2.0",
		Error: &Error{
			Code:    code,
			Message: message,
		},
		ID: id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC errors still return 200
	json.NewEncoder(w).Encode(resp)
}
