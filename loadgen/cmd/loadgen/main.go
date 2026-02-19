package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/hannanrazzaghi/wirefight/loadgen/internal/results"
	"github.com/hannanrazzaghi/wirefight/loadgen/internal/runner"
)

func main() {
	// Define flags
	var (
		protocol         = flag.String("protocol", "rest", "Protocol to test (rest, jsonrpc, grpc)")
		mode             = flag.String("mode", "cpu", "Workload mode (cpu, io)")
		workFactor       = flag.Int("work-factor", 100, "Work factor (iterations for cpu, ms for io)")
		payloadSize      = flag.Int("payload-size", 256, "Payload size in bytes")
		concurrency      = flag.Int("concurrency", 10, "Number of concurrent workers")
		duration         = flag.Duration("duration", 30*time.Second, "Test duration")
		warmup           = flag.Duration("warmup", 5*time.Second, "Warmup duration")
		output           = flag.String("output", "", "Output file for results (JSON)")
		httpAddr         = flag.String("http-addr", "http://localhost:8080", "HTTP address for REST/JSON-RPC")
		grpcAddr         = flag.String("grpc-addr", "localhost:9090", "gRPC address")
	)

	flag.Parse()

	// Validate flags
	if *protocol != "rest" && *protocol != "jsonrpc" && *protocol != "grpc" {
		log.Fatalf("Invalid protocol: %s (must be rest, jsonrpc, or grpc)", *protocol)
	}

	if *mode != "cpu" && *mode != "io" {
		log.Fatalf("Invalid mode: %s (must be cpu or io)", *mode)
	}

	if *output == "" {
		log.Fatal("Output file is required (use -output flag)")
	}

	// Create runner config
	cfg := runner.Config{
		Protocol:         *protocol,
		Mode:             *mode,
		WorkFactor:       *workFactor,
		PayloadSizeBytes: *payloadSize,
		Concurrency:      *concurrency,
		Duration:         *duration,
		WarmupDuration:   *warmup,
		HTTPAddr:         *httpAddr,
		GRPCAddr:         *grpcAddr,
	}

	// Print configuration
	log.Println("🔥 Wirefight Load Generator")
	log.Printf("   Protocol:    %s", cfg.Protocol)
	log.Printf("   Mode:        %s", cfg.Mode)
	log.Printf("   Work Factor: %d", cfg.WorkFactor)
	log.Printf("   Payload:     %d bytes", cfg.PayloadSizeBytes)
	log.Printf("   Concurrency: %d", cfg.Concurrency)
	log.Printf("   Duration:    %v", cfg.Duration)
	log.Printf("   Warmup:      %v", cfg.WarmupDuration)

	// Create runner
	r, err := runner.NewRunner(cfg)
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}
	defer r.Close()

	// Run load test
	ctx := context.Background()
	collector, err := r.Run(ctx)
	if err != nil {
		log.Fatalf("Load test failed: %v", err)
	}

	// Aggregate results
	stats := collector.Aggregate(cfg.Duration)

	// Create benchmark result
	result := results.BenchmarkResult{
		Protocol:         cfg.Protocol,
		Mode:             cfg.Mode,
		WorkFactor:       cfg.WorkFactor,
		PayloadSizeBytes: cfg.PayloadSizeBytes,
		Concurrency:      cfg.Concurrency,
		Duration:         cfg.Duration.Seconds(),
		WarmupDuration:   cfg.WarmupDuration.Seconds(),
		Stats:            stats,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
	}

	// Print summary
	log.Println("\n📊 Results:")
	log.Printf("   Total Requests:  %d", stats.TotalRequests)
	log.Printf("   Successful:      %d", stats.SuccessfulRequests)
	log.Printf("   Failed:          %d", stats.FailedRequests)
	log.Printf("   RPS:             %.2f", stats.RPS)
	log.Printf("   Mean Latency:    %.2f ms", stats.Mean)
	log.Printf("   p50:             %.2f ms", stats.P50)
	log.Printf("   p90:             %.2f ms", stats.P90)
	log.Printf("   p95:             %.2f ms", stats.P95)
	log.Printf("   p99:             %.2f ms", stats.P99)
	log.Printf("   Min:             %.2f ms", stats.Min)
	log.Printf("   Max:             %.2f ms", stats.Max)
	log.Printf("   Error Rate:      %.2f%%", stats.ErrorRate*100)

	// Save to file
	if err := results.SaveToFile(result, *output); err != nil {
		log.Fatalf("Failed to save results: %v", err)
	}

	log.Printf("\n✓ Results saved to: %s", *output)

	// Exit with error code if there were failures
	if stats.FailedRequests > 0 {
		os.Exit(1)
	}
}
