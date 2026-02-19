# Wirefight

Controlled localhost benchmarking lab comparing protocol overhead and behavior under load.

## Protocols Compared

1. **REST** - JSON over HTTP using chi router
2. **JSON-RPC 2.0** - Minimal implementation over HTTP
3. **gRPC** - Protocol buffers

All three execute **identical business logic** to measure pure protocol overhead.

## Architecture

- **Go Service**: Single binary exposing all three protocol endpoints
- **Go Load Generator**: Stress testing tool with configurable concurrency
- **Python Analytics**: Result parsing, statistical analysis, and visualization

## Project Structure

```
wirefight/
├── go-service/          # Service implementation
│   ├── cmd/server/      # Main entry point
│   └── internal/        # Protocol handlers + shared logic
├── loadgen/             # Load generator
│   ├── cmd/loadgen/     # CLI entry point
│   └── internal/        # Worker pool + result aggregation
├── api/proto/           # Protobuf definitions
├── benchmark/           # Python analytics
├── results/             # Benchmark output
└── Makefile             # Build and run automation
```

## Quick Start

```bash
# Build everything
make build

# Run service
make run-service

# Run benchmarks (in another terminal)
make benchmark

# Generate report
make analyze
```

## Workload Modes

- **cpu**: Deterministic SHA-256 hashing (work_factor = iterations)
- **io**: Simulated I/O via sleep (work_factor = sleep_ms)

## Requirements

- Go 1.21+
- Python 3.9+ with matplotlib
- protoc compiler

## Fairness Principles

- Same machine, same Go build
- Connection reuse (keep-alive, gRPC pooling)
- Warmup phase before measurement
- Debug logging disabled during load
- All assumptions documented
