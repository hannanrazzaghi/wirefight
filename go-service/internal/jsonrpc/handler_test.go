package jsonrpc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hannan/wirefight/go-service/internal/metrics"
)

func TestHandleRPC_Success(t *testing.T) {
	handler := NewHandler(metrics.NewCollector(), false)

	params := ComputeParams{
		RequestID:        "test-123",
		Mode:             "cpu",
		WorkFactor:       10,
		PayloadSizeBytes: 256,
	}

	paramsJSON, _ := json.Marshal(params)
	reqBody := Request{
		JSONRPC: "2.0",
		Method:  "compute",
		Params:  paramsJSON,
		ID:      "1",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleRPC(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error != nil {
		t.Errorf("Expected no error, got: %+v", resp.Error)
	}

	resultJSON, _ := json.Marshal(resp.Result)
	var result ComputeResult
	json.Unmarshal(resultJSON, &result)

	if result.RequestID != params.RequestID {
		t.Errorf("RequestID mismatch: got %s, want %s", result.RequestID, params.RequestID)
	}
	if result.Protocol != "jsonrpc" {
		t.Errorf("Protocol mismatch: got %s, want jsonrpc", result.Protocol)
	}
}

func TestHandleRPC_InvalidVersion(t *testing.T) {
	handler := NewHandler(metrics.NewCollector(), false)

	reqBody := Request{
		JSONRPC: "1.0",
		Method:  "compute",
		ID:      "1",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleRPC(w, req)

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Error == nil {
		t.Error("Expected error for invalid version")
	}
	if resp.Error.Code != InvalidRequest {
		t.Errorf("Expected error code %d, got %d", InvalidRequest, resp.Error.Code)
	}
}

func TestHandleRPC_MethodNotFound(t *testing.T) {
	handler := NewHandler(metrics.NewCollector(), false)

	reqBody := Request{
		JSONRPC: "2.0",
		Method:  "unknown",
		ID:      "1",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleRPC(w, req)

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Error == nil {
		t.Error("Expected error for unknown method")
	}
	if resp.Error.Code != MethodNotFound {
		t.Errorf("Expected error code %d, got %d", MethodNotFound, resp.Error.Code)
	}
}

func TestHandleRPC_InvalidParams(t *testing.T) {
	handler := NewHandler(metrics.NewCollector(), false)

	reqBody := Request{
		JSONRPC: "2.0",
		Method:  "compute",
		Params:  json.RawMessage(`{"request_id": "", "mode": "invalid"}`),
		ID:      "1",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleRPC(w, req)

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Error == nil {
		t.Error("Expected error for invalid params")
	}
}

func TestHandleRPC_ParseError(t *testing.T) {
	handler := NewHandler(metrics.NewCollector(), false)

	req := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.HandleRPC(w, req)

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Error == nil {
		t.Error("Expected parse error")
	}
	if resp.Error.Code != ParseError {
		t.Errorf("Expected error code %d, got %d", ParseError, resp.Error.Code)
	}
}

func TestHandleRPC_IOMode(t *testing.T) {
	handler := NewHandler(metrics.NewCollector(), false)

	params := ComputeParams{
		RequestID:  "test-io",
		Mode:       "io",
		WorkFactor: 5,
	}

	paramsJSON, _ := json.Marshal(params)
	reqBody := Request{
		JSONRPC: "2.0",
		Method:  "compute",
		Params:  paramsJSON,
		ID:      "2",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleRPC(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Error != nil {
		t.Errorf("Expected no error, got: %+v", resp.Error)
	}
}
