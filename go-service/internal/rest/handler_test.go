package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hannanrazzaghi/wirefight/go-service/internal/metrics"
)

func TestHandleCompute_Success(t *testing.T) {
	handler := NewHandler(metrics.NewCollector(), false)

	reqBody := ComputeRequest{
		RequestID:        "test-123",
		Mode:             "cpu",
		WorkFactor:       10,
		PayloadSizeBytes: 256,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/compute", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleCompute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp ComputeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.RequestID != reqBody.RequestID {
		t.Errorf("RequestID mismatch: got %s, want %s", resp.RequestID, reqBody.RequestID)
	}
	if resp.Protocol != "rest" {
		t.Errorf("Protocol mismatch: got %s, want rest", resp.Protocol)
	}
	if resp.ServerProcessingMs <= 0 {
		t.Error("ServerProcessingMs should be positive")
	}
}

func TestHandleCompute_MethodNotAllowed(t *testing.T) {
	handler := NewHandler(metrics.NewCollector(), false)

	req := httptest.NewRequest(http.MethodGet, "/v1/compute", nil)
	w := httptest.NewRecorder()

	handler.HandleCompute(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleCompute_InvalidJSON(t *testing.T) {
	handler := NewHandler(metrics.NewCollector(), false)

	req := httptest.NewRequest(http.MethodPost, "/v1/compute", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.HandleCompute(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleCompute_InvalidMode(t *testing.T) {
	handler := NewHandler(metrics.NewCollector(), false)

	reqBody := ComputeRequest{
		RequestID:  "test",
		Mode:       "invalid",
		WorkFactor: 10,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/compute", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleCompute(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleCompute_IOMode(t *testing.T) {
	handler := NewHandler(metrics.NewCollector(), false)

	reqBody := ComputeRequest{
		RequestID:        "test-io",
		Mode:             "io",
		WorkFactor:       10,
		PayloadSizeBytes: 0,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/compute", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleCompute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp ComputeResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Mode != "io" {
		t.Errorf("Mode mismatch: got %s, want io", resp.Mode)
	}
}
