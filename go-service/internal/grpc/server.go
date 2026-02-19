package grpc

import (
	"context"
	"log"

	"github.com/hannan/wirefight/api/proto"
	"github.com/hannan/wirefight/go-service/internal/logic"
	"github.com/hannan/wirefight/go-service/internal/metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the gRPC ComputeService
type Server struct {
	proto.UnimplementedComputeServiceServer
	metrics *metrics.Collector
	debug   bool
}

// NewServer creates a new gRPC server
func NewServer(m *metrics.Collector, debug bool) *Server {
	return &Server{
		metrics: m,
		debug:   debug,
	}
}

// Compute implements the Compute RPC method
func (s *Server) Compute(ctx context.Context, req *proto.ComputeRequest) (*proto.ComputeResponse, error) {
	s.metrics.IncrementGRPCRequests()

	// Convert from proto to internal logic request
	logicReq := logic.Request{
		RequestID:        req.RequestId,
		Mode:             logic.Mode(req.Mode),
		WorkFactor:       int(req.WorkFactor),
		PayloadSizeBytes: int(req.PayloadSizeBytes),
	}

	// Execute shared logic
	logicResp, err := logic.Execute(logicReq, "grpc")
	if err != nil {
		if s.debug {
			log.Printf("[gRPC] Logic execution failed: %v", err)
		}
		s.metrics.IncrementGRPCErrors()
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convert from internal logic response to proto
	resp := &proto.ComputeResponse{
		RequestId:          logicResp.RequestID,
		Mode:               string(logicResp.Mode),
		WorkFactor:         int32(logicResp.WorkFactor),
		PayloadSizeBytes:   int32(logicResp.PayloadSizeBytes),
		Result:             logicResp.Result,
		ServerProcessingMs: logicResp.ServerProcessingMs,
		Protocol:           logicResp.Protocol,
		Timestamp:          logicResp.Timestamp,
	}

	return resp, nil
}
