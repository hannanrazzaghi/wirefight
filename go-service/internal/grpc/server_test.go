package grpc

import (
	"context"
	"testing"

	"github.com/hannanrazzaghi/wirefight/api/proto"
	"github.com/hannanrazzaghi/wirefight/go-service/internal/metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCompute_Success(t *testing.T) {
	server := NewServer(metrics.NewCollector(), false)

	req := &proto.ComputeRequest{
		RequestId:        "test-123",
		Mode:             "cpu",
		WorkFactor:       10,
		PayloadSizeBytes: 256,
	}

	resp, err := server.Compute(context.Background(), req)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	if resp.RequestId != req.RequestId {
		t.Errorf("RequestID mismatch: got %s, want %s", resp.RequestId, req.RequestId)
	}
	if resp.Protocol != "grpc" {
		t.Errorf("Protocol mismatch: got %s, want grpc", resp.Protocol)
	}
	if resp.ServerProcessingMs <= 0 {
		t.Error("ServerProcessingMs should be positive")
	}
}

func TestCompute_InvalidMode(t *testing.T) {
	server := NewServer(metrics.NewCollector(), false)

	req := &proto.ComputeRequest{
		RequestId:  "test",
		Mode:       "invalid",
		WorkFactor: 10,
	}

	_, err := server.Compute(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for invalid mode")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected InvalidArgument code, got %v", st.Code())
	}
}

func TestCompute_MissingRequestID(t *testing.T) {
	server := NewServer(metrics.NewCollector(), false)

	req := &proto.ComputeRequest{
		Mode:       "cpu",
		WorkFactor: 10,
	}

	_, err := server.Compute(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for missing request_id")
	}
}

func TestCompute_IOMode(t *testing.T) {
	server := NewServer(metrics.NewCollector(), false)

	req := &proto.ComputeRequest{
		RequestId:  "test-io",
		Mode:       "io",
		WorkFactor: 10,
	}

	resp, err := server.Compute(context.Background(), req)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	if resp.Mode != "io" {
		t.Errorf("Mode mismatch: got %s, want io", resp.Mode)
	}
}

func TestCompute_WithContext(t *testing.T) {
	server := NewServer(metrics.NewCollector(), false)

	ctx := context.Background()
	req := &proto.ComputeRequest{
		RequestId:        "test-ctx",
		Mode:             "cpu",
		WorkFactor:       5,
		PayloadSizeBytes: 100,
	}

	resp, err := server.Compute(ctx, req)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
}
