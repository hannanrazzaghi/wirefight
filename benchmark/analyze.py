#!/usr/bin/env python3

import glob
import json
import os
import platform
import socket
import subprocess
from collections import defaultdict
from datetime import UTC, datetime
from pathlib import Path

from charts import render_mode_charts

ROOT = Path(__file__).resolve().parent.parent
RESULTS_DIR = ROOT / "results"
CHARTS_DIR = RESULTS_DIR / "charts"
TEMPLATE_FILE = ROOT / "benchmark" / "report_template.md"
SUMMARY_FILE = RESULTS_DIR / "BENCHMARK_SUMMARY.md"
REPORT_FILE = RESULTS_DIR / "BENCHMARK_REPORT.md"


def load_rows(results_dir: Path):
    rows = []
    for file_path in glob.glob(str(results_dir / "*.json")):
        with open(file_path, "r", encoding="utf-8") as file_handle:
            payload = json.load(file_handle)
        stats = payload.get("stats", {})
        rows.append(
            {
                "file": os.path.basename(file_path),
                "protocol": payload.get("protocol"),
                "mode": payload.get("mode"),
                "work_factor": payload.get("work_factor"),
                "payload_size_bytes": payload.get("payload_size_bytes"),
                "concurrency": payload.get("concurrency"),
                "duration_seconds": payload.get("duration"),
                "rps": stats.get("rps", 0.0),
                "p50": stats.get("p50", 0.0),
                "p90": stats.get("p90", 0.0),
                "p95": stats.get("p95", 0.0),
                "p99": stats.get("p99", 0.0),
                "mean": stats.get("mean", 0.0),
                "max": stats.get("max", 0.0),
                "error_rate": stats.get("error_rate", 0.0),
                "timestamp": payload.get("timestamp", ""),
            }
        )
    return rows


def write_summary(rows, output_file: Path):
    grouped = defaultdict(list)
    for row in rows:
        grouped[(row["mode"], row["payload_size_bytes"], row["work_factor"])].append(row)

    lines = ["# Benchmark Summary", ""]
    for key in sorted(grouped.keys()):
        mode, payload_size, work_factor = key
        lines.append(f"## mode={mode}, payload={payload_size}, work_factor={work_factor}")
        lines.append("")
        lines.append("| protocol | concurrency | rps | p50 | p95 | p99 | mean | error_rate |")
        lines.append("|---|---:|---:|---:|---:|---:|---:|---:|")

        for row in sorted(grouped[key], key=lambda item: (item["protocol"], item["concurrency"])):
            lines.append(
                f"| {row['protocol']} | {row['concurrency']} | {row['rps']:.2f} | {row['p50']:.2f} | "
                f"{row['p95']:.2f} | {row['p99']:.2f} | {row['mean']:.2f} | {row['error_rate'] * 100:.2f}% |"
            )
        lines.append("")

    output_file.write_text("\n".join(lines), encoding="utf-8")


def generate_observations(rows):
    by_mode = defaultdict(list)
    for row in rows:
        by_mode[row["mode"]].append(row)

    bullets = []
    for mode, mode_rows in sorted(by_mode.items()):
        best_rps = max(mode_rows, key=lambda item: item["rps"])
        best_p95 = min(mode_rows, key=lambda item: item["p95"])
        lowest_error = min(mode_rows, key=lambda item: item["error_rate"])
        bullets.append(
            f"- {mode.upper()}: highest throughput from `{best_rps['protocol']}` at concurrency={best_rps['concurrency']} "
            f"(rps={best_rps['rps']:.2f})."
        )
        bullets.append(
            f"- {mode.upper()}: lowest p95 latency from `{best_p95['protocol']}` at concurrency={best_p95['concurrency']} "
            f"(p95={best_p95['p95']:.2f} ms)."
        )
        bullets.append(
            f"- {mode.upper()}: lowest error rate from `{lowest_error['protocol']}` at concurrency={lowest_error['concurrency']} "
            f"(error={lowest_error['error_rate'] * 100:.2f}%)."
        )

    return "\n".join(bullets)


def generate_tradeoffs(rows):
    protocols = sorted({row["protocol"] for row in rows})
    lines = []
    for protocol in protocols:
        protocol_rows = [row for row in rows if row["protocol"] == protocol]
        avg_rps = sum(r["rps"] for r in protocol_rows) / len(protocol_rows)
        avg_p95 = sum(r["p95"] for r in protocol_rows) / len(protocol_rows)
        avg_err = (sum(r["error_rate"] for r in protocol_rows) / len(protocol_rows)) * 100.0
        lines.append(
            f"- `{protocol}`: average rps={avg_rps:.2f}, average p95={avg_p95:.2f} ms, average error={avg_err:.2f}%."
        )
    return "\n".join(lines)


def read_go_version():
    try:
        out = subprocess.check_output(["go", "version"], stderr=subprocess.STDOUT, text=True)
        return out.strip()
    except Exception:
        return "unknown"


def write_report(rows, output_file: Path):
    template = TEMPLATE_FILE.read_text(encoding="utf-8")
    cpu_work_factors = sorted({row["work_factor"] for row in rows if row["mode"] == "cpu"})
    io_work_factors = sorted({row["work_factor"] for row in rows if row["mode"] == "io"})

    report = template.format(
        date=datetime.now(UTC).strftime("%Y-%m-%d"),
        host=socket.gethostname(),
        go_version=read_go_version(),
        python_version=platform.python_version(),
        cpu_work_factors=", ".join(str(x) for x in cpu_work_factors) or "n/a",
        io_work_factors=", ".join(str(x) for x in io_work_factors) or "n/a",
        observations=generate_observations(rows),
        tradeoffs=generate_tradeoffs(rows),
    )
    output_file.write_text(report, encoding="utf-8")


def main():
    rows = load_rows(RESULTS_DIR)
    if not rows:
        raise SystemExit("No result files found in results/")

    grouped = defaultdict(list)
    for row in rows:
        grouped[row["mode"]].append(row)

    generated = render_mode_charts(grouped, str(CHARTS_DIR))
    write_summary(rows, SUMMARY_FILE)
    write_report(rows, REPORT_FILE)

    print(f"Loaded {len(rows)} result files")
    print(f"Generated {len(generated)} chart files in {CHARTS_DIR}")
    print(f"Wrote summary: {SUMMARY_FILE}")
    print(f"Wrote report: {REPORT_FILE}")


if __name__ == "__main__":
    main()
