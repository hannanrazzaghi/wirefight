package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hannan/wirefight/api/proto"
	"github.com/hannan/wirefight/go-service/internal/config"
	grpcHandler "github.com/hannan/wirefight/go-service/internal/grpc"
	"github.com/hannan/wirefight/go-service/internal/jsonrpc"
	"github.com/hannan/wirefight/go-service/internal/metrics"
	"github.com/hannan/wirefight/go-service/internal/rest"
	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize metrics collector
	metricsCollector := metrics.NewCollector()

	// Setup HTTP server for REST and JSON-RPC
	httpServer := setupHTTPServer(cfg, metricsCollector)

	// Setup gRPC server
	grpcServer := setupGRPCServer(cfg, metricsCollector)

	// Setup pprof server if enabled
	var pprofServer *http.Server
	if cfg.EnablePProf {
		pprofServer = setupPProfServer(cfg)
	}

	// Start servers
	httpErrChan := make(chan error, 1)
	grpcErrChan := make(chan error, 1)
	pprofErrChan := make(chan error, 1)

	go startHTTPServer(httpServer, cfg.HTTPPort, httpErrChan)
	go startGRPCServer(grpcServer, cfg.GRPCPort, grpcErrChan)
	if pprofServer != nil {
		go startPProfServer(pprofServer, cfg.PProfPort, pprofErrChan)
	}

	log.Println("🚀 Wirefight service started")
	log.Printf("   REST:     http://localhost:%d/v1/compute", cfg.HTTPPort)
	log.Printf("   JSON-RPC: http://localhost:%d/rpc", cfg.HTTPPort)
	log.Printf("   gRPC:     localhost:%d", cfg.GRPCPort)
	log.Printf("   Health:   http://localhost:%d/health", cfg.HTTPPort)
	log.Printf("   Metrics:  http://localhost:%d/metrics", cfg.HTTPPort)
	if cfg.EnablePProf {
		log.Printf("   PProf:    http://localhost:%d/debug/pprof", cfg.PProfPort)
	}

	// Wait for shutdown signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-httpErrChan:
		log.Fatalf("HTTP server error: %v", err)
	case err := <-grpcErrChan:
		log.Fatalf("gRPC server error: %v", err)
	case err := <-pprofErrChan:
		log.Fatalf("PProf server error: %v", err)
	case <-shutdown:
		log.Println("\n🛑 Shutting down gracefully...")

		// Graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}

		grpcServer.GracefulStop()

		if pprofServer != nil {
			if err := pprofServer.Shutdown(ctx); err != nil {
				log.Printf("PProf server shutdown error: %v", err)
			}
		}

		log.Println("✓ Shutdown complete")
	}
}

func setupHTTPServer(cfg *config.Config, m *metrics.Collector) *http.Server {
	r := chi.NewRouter()

	// Middleware
	if !cfg.EnableDebugLog {
		// Only add request logging in debug mode
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)

	// Create handlers
	restHandler := rest.NewHandler(m, cfg.EnableDebugLog)
	jsonrpcHandler := jsonrpc.NewHandler(m, cfg.EnableDebugLog)

	// REST endpoint
	r.Post("/v1/compute", restHandler.HandleCompute)

	// JSON-RPC endpoint
	r.Post("/rpc", jsonrpcHandler.HandleRPC)

	// Observability endpoints
	r.Get("/health", metrics.HealthHandler())
	r.Get("/metrics", m.HTTPHandler())

	return &http.Server{
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func setupGRPCServer(cfg *config.Config, m *metrics.Collector) *grpc.Server {
	grpcServer := grpc.NewServer()
	computeServer := grpcHandler.NewServer(m, cfg.EnableDebugLog)
	proto.RegisterComputeServiceServer(grpcServer, computeServer)
	return grpcServer
}

func setupPProfServer(cfg *config.Config) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return &http.Server{
		Handler: mux,
	}
}

func startHTTPServer(server *http.Server, port int, errChan chan<- error) {
	server.Addr = fmt.Sprintf(":%d", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		errChan <- fmt.Errorf("HTTP server error: %w", err)
	}
}

func startGRPCServer(server *grpc.Server, port int, errChan chan<- error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		errChan <- fmt.Errorf("gRPC listen error: %w", err)
		return
	}

	if err := server.Serve(lis); err != nil {
		errChan <- fmt.Errorf("gRPC serve error: %w", err)
	}
}

func startPProfServer(server *http.Server, port int, errChan chan<- error) {
	server.Addr = fmt.Sprintf(":%d", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		errChan <- fmt.Errorf("PProf server error: %w", err)
	}
}
