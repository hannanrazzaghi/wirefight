.PHONY: help build clean test run-service benchmark benchmark-quick benchmark-matrix analyze proto install-tools

# Default target
help:
	@echo "Wirefight - Protocol Benchmarking System"
	@echo ""
	@echo "Available targets:"
	@echo "  make build          - Build service and loadgen binaries"
	@echo "  make proto          - Generate protobuf code"
	@echo "  make test           - Run all tests"
	@echo "  make run-service    - Start the benchmark service"
	@echo "  make benchmark      - Run full benchmark suite"
	@echo "  make benchmark-quick - Run quick benchmark"
	@echo "  make benchmark-matrix- Run full protocol/mode/matrix sweep"
	@echo "  make analyze        - Analyze benchmark results"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make install-tools  - Install required tools"

# Build targets
build: build-service build-loadgen

build-service:
	@echo "Building service..."
	cd go-service && go build -o server ./cmd/server
	@echo "✓ Service built: go-service/server"

build-loadgen:
	@echo "Building loadgen..."
	cd loadgen && go build -o loadgen ./cmd/loadgen
	@echo "✓ Loadgen built: loadgen/loadgen"

# Proto generation
proto:
	@echo "Generating protobuf code..."
	./scripts/generate-proto.sh

# Testing
test:
	@echo "Running go-service tests..."
	cd go-service && go test ./...
	@echo ""
	@echo "Running loadgen tests..."
	cd loadgen && go test ./...
	@echo "✓ All tests passed"

# Run service
run-service: build-service
	@echo "Starting wirefight service..."
	cd go-service && ./server

# Benchmark targets
benchmark: build benchmark-cpu benchmark-io
	@echo ""
	@echo "✓ Full benchmark suite complete"
	@echo "Run 'make analyze' to generate charts and report"

benchmark-matrix: build-loadgen
	@echo "Running full benchmark matrix (this can take a long time)..."
	@mkdir -p results
	@for protocol in rest jsonrpc grpc; do \
		for mode in cpu io; do \
			for conc in 1 10 50 100 250; do \
				for payload in 256 4096 65536; do \
					if [ "$$mode" = "cpu" ]; then wfs="100 1000"; else wfs="5 20"; fi; \
					for wf in $$wfs; do \
						echo "[matrix] $$protocol $$mode c=$$conc payload=$$payload wf=$$wf"; \
						./loadgen/loadgen -protocol=$$protocol -mode=$$mode -concurrency=$$conc -duration=20s -warmup=5s -work-factor=$$wf -payload-size=$$payload -output=results/$$protocol_$$mode_c$$conc_p$$payload_wf$$wf.json || true; \
					done; \
				done; \
			done; \
		done; \
	done
	@echo "✓ Matrix benchmark complete"

benchmark-quick: build
	@echo "Running quick benchmark (REST only, low concurrency)..."
	@mkdir -p results
	./loadgen/loadgen -protocol=rest -mode=cpu -concurrency=10 -duration=10s -warmup=2s -work-factor=100 -payload-size=256 -output=results/quick_rest_cpu.json
	@echo "✓ Quick benchmark complete"

benchmark-cpu: build-loadgen
	@echo "Running CPU benchmarks..."
	@mkdir -p results
	# REST - CPU mode, various concurrency levels
	@echo "  Testing REST..."
	./loadgen/loadgen -protocol=rest -mode=cpu -concurrency=1 -duration=30s -warmup=5s -work-factor=100 -payload-size=256 -output=results/rest_cpu_c1_p256_wf100.json
	./loadgen/loadgen -protocol=rest -mode=cpu -concurrency=10 -duration=30s -warmup=5s -work-factor=100 -payload-size=256 -output=results/rest_cpu_c10_p256_wf100.json
	./loadgen/loadgen -protocol=rest -mode=cpu -concurrency=50 -duration=30s -warmup=5s -work-factor=100 -payload-size=256 -output=results/rest_cpu_c50_p256_wf100.json
	./loadgen/loadgen -protocol=rest -mode=cpu -concurrency=100 -duration=30s -warmup=5s -work-factor=100 -payload-size=256 -output=results/rest_cpu_c100_p256_wf100.json
	# JSON-RPC - CPU mode
	@echo "  Testing JSON-RPC..."
	./loadgen/loadgen -protocol=jsonrpc -mode=cpu -concurrency=1 -duration=30s -warmup=5s -work-factor=100 -payload-size=256 -output=results/jsonrpc_cpu_c1_p256_wf100.json
	./loadgen/loadgen -protocol=jsonrpc -mode=cpu -concurrency=10 -duration=30s -warmup=5s -work-factor=100 -payload-size=256 -output=results/jsonrpc_cpu_c10_p256_wf100.json
	./loadgen/loadgen -protocol=jsonrpc -mode=cpu -concurrency=50 -duration=30s -warmup=5s -work-factor=100 -payload-size=256 -output=results/jsonrpc_cpu_c50_p256_wf100.json
	./loadgen/loadgen -protocol=jsonrpc -mode=cpu -concurrency=100 -duration=30s -warmup=5s -work-factor=100 -payload-size=256 -output=results/jsonrpc_cpu_c100_p256_wf100.json
	# gRPC - CPU mode
	@echo "  Testing gRPC..."
	./loadgen/loadgen -protocol=grpc -mode=cpu -concurrency=1 -duration=30s -warmup=5s -work-factor=100 -payload-size=256 -output=results/grpc_cpu_c1_p256_wf100.json
	./loadgen/loadgen -protocol=grpc -mode=cpu -concurrency=10 -duration=30s -warmup=5s -work-factor=100 -payload-size=256 -output=results/grpc_cpu_c10_p256_wf100.json
	./loadgen/loadgen -protocol=grpc -mode=cpu -concurrency=50 -duration=30s -warmup=5s -work-factor=100 -payload-size=256 -output=results/grpc_cpu_c50_p256_wf100.json
	./loadgen/loadgen -protocol=grpc -mode=cpu -concurrency=100 -duration=30s -warmup=5s -work-factor=100 -payload-size=256 -output=results/grpc_cpu_c100_p256_wf100.json

benchmark-io: build-loadgen
	@echo "Running I/O benchmarks..."
	@mkdir -p results
	# REST - IO mode
	@echo "  Testing REST..."
	./loadgen/loadgen -protocol=rest -mode=io -concurrency=10 -duration=30s -warmup=5s -work-factor=10 -payload-size=256 -output=results/rest_io_c10_p256_wf10.json
	./loadgen/loadgen -protocol=rest -mode=io -concurrency=50 -duration=30s -warmup=5s -work-factor=10 -payload-size=256 -output=results/rest_io_c50_p256_wf10.json
	./loadgen/loadgen -protocol=rest -mode=io -concurrency=100 -duration=30s -warmup=5s -work-factor=10 -payload-size=256 -output=results/rest_io_c100_p256_wf10.json
	# JSON-RPC - IO mode
	@echo "  Testing JSON-RPC..."
	./loadgen/loadgen -protocol=jsonrpc -mode=io -concurrency=10 -duration=30s -warmup=5s -work-factor=10 -payload-size=256 -output=results/jsonrpc_io_c10_p256_wf10.json
	./loadgen/loadgen -protocol=jsonrpc -mode=io -concurrency=50 -duration=30s -warmup=5s -work-factor=10 -payload-size=256 -output=results/jsonrpc_io_c50_p256_wf10.json
	./loadgen/loadgen -protocol=jsonrpc -mode=io -concurrency=100 -duration=30s -warmup=5s -work-factor=10 -payload-size=256 -output=results/jsonrpc_io_c100_p256_wf10.json
	# gRPC - IO mode
	@echo "  Testing gRPC..."
	./loadgen/loadgen -protocol=grpc -mode=io -concurrency=10 -duration=30s -warmup=5s -work-factor=10 -payload-size=256 -output=results/grpc_io_c10_p256_wf10.json
	./loadgen/loadgen -protocol=grpc -mode=io -concurrency=50 -duration=30s -warmup=5s -work-factor=10 -payload-size=256 -output=results/grpc_io_c50_p256_wf10.json
	./loadgen/loadgen -protocol=grpc -mode=io -concurrency=100 -duration=30s -warmup=5s -work-factor=10 -payload-size=256 -output=results/grpc_io_c100_p256_wf10.json

# Analysis
analyze:
	@echo "Analyzing benchmark results..."
	python3 benchmark/analyze.py
	@echo "✓ Analysis complete. Check results/charts/, results/BENCHMARK_SUMMARY.md, and results/BENCHMARK_REPORT.md"

# Cleanup
clean:
	@echo "Cleaning build artifacts..."
	rm -f go-service/server
	rm -f loadgen/loadgen
	rm -rf results/*.json
	rm -rf results/charts/*.png
	rm -f results/BENCHMARK_SUMMARY.md
	@echo "✓ Clean complete"

# Tool installation
install-tools:
	@echo "Installing required tools..."
	@echo "Checking Go..."
	@which go > /dev/null || (echo "Go not found. Install from https://golang.org/" && exit 1)
	@echo "✓ Go found"
	@echo ""
	@echo "Checking protoc..."
	@which protoc > /dev/null || (echo "protoc not found. Install from https://grpc.io/docs/protoc-installation/" && exit 1)
	@echo "✓ protoc found"
	@echo ""
	@echo "Installing Go protobuf tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "✓ Go protobuf tools installed"
	@echo ""
	@echo "Checking Python..."
	@which python3 > /dev/null || (echo "Python 3 not found" && exit 1)
	@echo "✓ Python found"
	@echo ""
	@echo "Installing Python dependencies..."
	python3 -m pip install -r benchmark/requirements.txt
	@echo "✓ Python dependencies installed"
	@echo ""
	@echo "✓ All tools installed successfully"
