# Latest Matrix Report (Accelerated)

- Run type: accelerated full matrix for in-session completion
- Coverage: protocols=3, modes=2, concurrency=5, payloads=3, work_factors=2
- Total matrix points: 180
- Warmup/Duration per point: 1s warmup + 2s measurement

## High-level findings

### CPU mode

- Best average throughput: `rest` (10026.48 RPS)
- Best average p95 latency: `grpc` (98.84 ms)
- Lowest average error rate: `rest` (2.78%)

| Protocol | Avg RPS | Avg p95 (ms) | Avg p99 (ms) | Avg error (%) |
|---|---:|---:|---:|---:|
| rest | 10026.48 | 123.02 | 155.17 | 2.78 |
| jsonrpc | 9516.33 | 129.95 | 160.86 | 2.92 |
| grpc | 9652.08 | 98.84 | 197.29 | 9.57 |

### IO mode

- Best average throughput: `rest` (7904.30 RPS)
- Best average p95 latency: `grpc` (16.99 ms)
- Lowest average error rate: `rest` (0.71%)

| Protocol | Avg RPS | Avg p95 (ms) | Avg p99 (ms) | Avg error (%) |
|---|---:|---:|---:|---:|
| rest | 7904.30 | 17.68 | 21.71 | 0.71 |
| jsonrpc | 7890.87 | 17.94 | 22.69 | 0.76 |
| grpc | 7052.75 | 16.99 | 80.67 | 7.63 |

## Notes

- These results prioritize breadth (full matrix) over long steady-state duration.
- For publication-grade measurements, run the standard `make benchmark-matrix` target with longer windows.
- Matrix run artifacts are written under `results/matrixfast_*.json`.
