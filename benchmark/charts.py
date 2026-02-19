#!/usr/bin/env python3

import os
from typing import Dict, List

import matplotlib

matplotlib.use("Agg")
import matplotlib.pyplot as plt


def render_mode_charts(grouped_rows: Dict[str, List[dict]], output_dir: str) -> List[str]:
    os.makedirs(output_dir, exist_ok=True)
    generated = []

    for mode, rows in grouped_rows.items():
        protocols = sorted({row["protocol"] for row in rows})
        payloads = sorted({row["payload_size_bytes"] for row in rows})

        for payload in payloads:
            payload_rows = [r for r in rows if r["payload_size_bytes"] == payload]
            fig, axes = plt.subplots(2, 2, figsize=(15, 10))
            fig.suptitle(f"{mode.upper()} mode, payload={payload}B", fontsize=14)

            _plot_metric(axes[0][0], payload_rows, protocols, "rps", "Throughput vs Concurrency", "RPS")
            _plot_metric(axes[0][1], payload_rows, protocols, "p95", "p95 Latency vs Concurrency", "Latency (ms)")
            _plot_metric(axes[1][0], payload_rows, protocols, "p99", "p99 Latency vs Concurrency", "Latency (ms)")

            err_rows = []
            for row in payload_rows:
                copied = dict(row)
                copied["error_rate"] = copied["error_rate"] * 100.0
                err_rows.append(copied)
            _plot_metric(axes[1][1], err_rows, protocols, "error_rate", "Error Rate vs Concurrency", "Error rate (%)")

            plt.tight_layout()
            out_file = os.path.join(output_dir, f"{mode}_payload_{payload}.png")
            fig.savefig(out_file, dpi=150)
            plt.close(fig)
            generated.append(out_file)

    return generated


def _plot_metric(ax, rows: List[dict], protocols: List[str], metric_key: str, title: str, ylabel: str) -> None:
    markers = ["o", "s", "^", "d", "x"]
    for idx, protocol in enumerate(protocols):
        points = sorted(
            [row for row in rows if row["protocol"] == protocol],
            key=lambda item: item["concurrency"],
        )
        if not points:
            continue
        x = [p["concurrency"] for p in points]
        y = [p[metric_key] for p in points]
        ax.plot(x, y, marker=markers[idx % len(markers)], linewidth=2, label=protocol)

    ax.set_title(title)
    ax.set_xlabel("Concurrency")
    ax.set_ylabel(ylabel)
    ax.grid(True, alpha=0.3)
    ax.legend()
