package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/hannanrazzaghi/wirefight/api/proto"
	"github.com/hannanrazzaghi/wirefight/loadgen/internal/results"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config contains the load test configuration
type Config struct {
	Protocol         string
	Mode             string
	WorkFactor       int
	PayloadSizeBytes int
	Concurrency      int
	Duration         time.Duration
	WarmupDuration   time.Duration
	HTTPAddr         string
	GRPCAddr         string
}

// Runner executes load tests
type Runner struct {
	config     Config
	httpClient *http.Client
	grpcConn   *grpc.ClientConn
	grpcClient proto.ComputeServiceClient
}

// NewRunner creates a new load test runner
func NewRunner(cfg Config) (*Runner, error) {
	r := &Runner{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        cfg.Concurrency * 2,
				MaxIdleConnsPerHost: cfg.Concurrency * 2,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}

	// Setup gRPC connection if needed
	if cfg.Protocol == "grpc" {
		conn, err := grpc.NewClient(cfg.GRPCAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(10*1024*1024),
				grpc.MaxCallSendMsgSize(10*1024*1024),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to gRPC: %w", err)
		}
		r.grpcConn = conn
		r.grpcClient = proto.NewComputeServiceClient(conn)
	}

	return r, nil
}

// Close releases resources
func (r *Runner) Close() error {
	if r.grpcConn != nil {
		return r.grpcConn.Close()
	}
	return nil
}

// Run executes the load test
func (r *Runner) Run(ctx context.Context) (*results.Collector, error) {
	log.Printf("Starting warmup phase (%v)...", r.config.WarmupDuration)
	if err := r.runPhase(ctx, r.config.WarmupDuration, nil); err != nil {
		return nil, fmt.Errorf("warmup failed: %w", err)
	}

	log.Printf("Starting measurement phase (%v)...", r.config.Duration)
	collector := results.NewCollector()
	if err := r.runPhase(ctx, r.config.Duration, collector); err != nil {
		return nil, fmt.Errorf("measurement failed: %w", err)
	}

	return collector, nil
}

// runPhase runs a single phase (warmup or measurement)
func (r *Runner) runPhase(ctx context.Context, duration time.Duration, collector *results.Collector) error {
	var wg sync.WaitGroup
	resultChan := make(chan results.RequestResult, r.config.Concurrency*10)

	// Start result collector goroutine if collecting results
	var collectorWg sync.WaitGroup
	if collector != nil {
		collectorWg.Add(1)
		go func() {
			defer collectorWg.Done()
			for result := range resultChan {
				collector.Add(result)
			}
		}()
	}

	// Create context with timeout
	phaseCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	// Start worker goroutines
	for i := 0; i < r.config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			r.worker(phaseCtx, workerID, resultChan, collector != nil)
		}(i)
	}

	// Wait for all workers to finish
	wg.Wait()
	close(resultChan)

	// Wait for result collection to complete
	if collector != nil {
		collectorWg.Wait()
	}

	return nil
}

// worker is a single worker that sends requests continuously
func (r *Runner) worker(ctx context.Context, workerID int, resultChan chan<- results.RequestResult, collect bool) {
	requestID := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			requestID++
			reqID := fmt.Sprintf("worker-%d-req-%d", workerID, requestID)

			start := time.Now()
			err := r.sendRequest(ctx, reqID)
			end := time.Now()

			if collect {
				result := results.RequestResult{
					StartTime: start,
					EndTime:   end,
					LatencyMs: float64(end.Sub(start).Microseconds()) / 1000.0,
					Success:   err == nil,
				}
				if err != nil {
					result.Error = err.Error()
				}
				resultChan <- result
			}
		}
	}
}

// sendRequest sends a single request based on the protocol
func (r *Runner) sendRequest(ctx context.Context, requestID string) error {
	switch r.config.Protocol {
	case "rest":
		return r.sendRESTRequest(ctx, requestID)
	case "jsonrpc":
		return r.sendJSONRPCRequest(ctx, requestID)
	case "grpc":
		return r.sendGRPCRequest(ctx, requestID)
	default:
		return fmt.Errorf("unknown protocol: %s", r.config.Protocol)
	}
}

// sendRESTRequest sends a REST request
func (r *Runner) sendRESTRequest(ctx context.Context, requestID string) error {
	reqBody := map[string]interface{}{
		"request_id":         requestID,
		"mode":               r.config.Mode,
		"work_factor":        r.config.WorkFactor,
		"payload_size_bytes": r.config.PayloadSizeBytes,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.config.HTTPAddr+"/v1/compute", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read and discard body to enable connection reuse
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

// sendJSONRPCRequest sends a JSON-RPC request
func (r *Runner) sendJSONRPCRequest(ctx context.Context, requestID string) error {
	params := map[string]interface{}{
		"request_id":         requestID,
		"mode":               r.config.Mode,
		"work_factor":        r.config.WorkFactor,
		"payload_size_bytes": r.config.PayloadSizeBytes,
	}

	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "compute",
		"params":  params,
		"id":      requestID,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.config.HTTPAddr+"/rpc", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read and discard body to enable connection reuse
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

// sendGRPCRequest sends a gRPC request
func (r *Runner) sendGRPCRequest(ctx context.Context, requestID string) error {
	req := &proto.ComputeRequest{
		RequestId:        requestID,
		Mode:             r.config.Mode,
		WorkFactor:       int32(r.config.WorkFactor),
		PayloadSizeBytes: int32(r.config.PayloadSizeBytes),
	}

	_, err := r.grpcClient.Compute(ctx, req)
	return err
}
