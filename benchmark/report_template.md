# Wirefight Benchmark Report

## Methodology

- Localhost-only execution.
- Same server binary for all protocols.
- Warmup phase excluded from measured metrics.
- Load generator reuses HTTP keep-alive connections and gRPC connections.
- Shared business logic path is identical across REST, JSON-RPC, and gRPC.

## Environment

- Date: {date}
- Host: {host}
- Go version: {go_version}
- Python version: {python_version}

## Experiment Matrix

- Protocols: `rest`, `jsonrpc`, `grpc`
- Modes: `cpu`, `io`
- Concurrency: `1, 10, 50, 100, 250`
- Payload sizes (bytes): `256, 4096, 65536`
- Work factors:
  - CPU: `{cpu_work_factors}`
  - IO: `{io_work_factors}`

## Observations

{observations}

## Protocol Trade-offs

{tradeoffs}

## Artifacts

- Charts directory: `results/charts/`
- Raw results: `results/*.json`
- Summary table: `results/BENCHMARK_SUMMARY.md`
