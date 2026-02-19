# Benchmark Analysis

Python tools for analyzing wirefight benchmark results.

## Setup

```bash
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
```

## Usage

1. Run benchmarks to generate JSON results in `../results/`
2. Analyze results:

```bash
python analyze.py
```

This will:
- Parse all JSON files in the results directory
- Generate comparative charts in `results/charts/`
- Create `results/BENCHMARK_SUMMARY.md` with tabulated results
- Create `results/BENCHMARK_REPORT.md` from `report_template.md`

## Output

### Charts Generated

For each combination of (mode, payload_size, work_factor):
- Throughput vs Concurrency
- p95 Latency vs Concurrency
- p99 Latency vs Concurrency
- Error Rate vs Concurrency

### Summary Report

Markdown file with tables comparing all protocols across different concurrency levels.
