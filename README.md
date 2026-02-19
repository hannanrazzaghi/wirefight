# wirefight

Localhost-only benchmarking lab to compare protocol overhead across:

- REST (JSON over HTTP)
- JSON-RPC 2.0 (over HTTP)
- gRPC (protobuf)

All three call the same internal business logic so the comparison isolates protocol behavior.

## Assumptions

- Benchmarks run on a single machine (localhost only).
- Server and load generator use the same Go toolchain/build.
- Debug logs are disabled during benchmark runs.
- HTTP keep-alive and gRPC connection reuse stay enabled.
- Warmup traffic is excluded from measured metrics.

## Repository Layout

```text
wirefight/
├── go-service/
│   ├── cmd/server/main.go
│   └── internal/
│       ├── config/
│       ├── grpc/
│       ├── jsonrpc/
│       ├── logic/
│       ├── metrics/
│       └── rest/
├── loadgen/
│   ├── cmd/loadgen/main.go
│   └── internal/
│       ├── results/
│       ├── runner/
│       └── scenarios/
├── api/proto/compute.proto
├── benchmark/
│   ├── analyze.py
│   ├── charts.py
│   └── report_template.md
├── results/
├── scripts/generate-proto.sh
└── Makefile
```

## Service Endpoints

- REST: `POST /v1/compute`
- JSON-RPC: `POST /rpc` (`method = "compute"`)
- gRPC: `ComputeService.Compute`
- Health: `GET /health`
- Metrics: `GET /metrics`
- pprof (separate port): `GET /debug/pprof/`

## Request/Response Semantics

Request fields:

- `request_id`
- `mode` (`cpu` | `io`)
- `work_factor`
- `payload_size_bytes`

Response fields:

- `request_id`
- `mode`
- `work_factor`
- `payload_size_bytes`
- `result`
- `server_processing_ms`
- `protocol`
- `timestamp`

## Workload Modes

- `cpu`: deterministic SHA-256 loop (`iterations = work_factor`)
- `io`: simulated I/O via sleep (`sleep_ms = work_factor`)

## Benchmark Matrix

Supported matrix sweep target (`make benchmark-matrix`):

- Protocols: `rest`, `jsonrpc`, `grpc`
- Modes: `cpu`, `io`
- Concurrency: `1, 10, 50, 100, 250`
- Payload sizes: `256, 4096, 65536` bytes
- Work factors:
  - CPU: `100, 1000`
  - IO: `5, 20`

## Prerequisites

- Go 1.21+
- Python 3.9+
- `protoc`
- `protoc-gen-go` and `protoc-gen-go-grpc`

## Quick Start

```bash
make install-tools
make proto
make build
make test
```

Start service:

```bash
make run-service
```

In a second terminal run benchmarks:

```bash
make benchmark-quick
# or
make benchmark-matrix
```

Generate analysis and report:

```bash
make analyze
```

Outputs:

- Raw results: `results/*.json`
- Charts: `results/charts/*.png`
- Summary: `results/BENCHMARK_SUMMARY.md`
- Report: `results/BENCHMARK_REPORT.md`

## Commit Discipline

Recommended conventional commit prefixes:

- `feat:` new functionality
- `fix:` bug fix
- `refactor:` code restructuring
- `test:` test changes
- `docs:` documentation
- `perf:` performance-focused changes
- `chore:` maintenance and tooling
