#!/usr/bin/env python3
"""
Wirefight Benchmark Analysis Tool

Parses JSON benchmark results and generates comparative analysis with charts.
"""

import json
import glob
import os
import sys
from dataclasses import dataclass
from typing import List, Dict
import matplotlib.pyplot as plt
import matplotlib
matplotlib.use('Agg')  # Non-interactive backend


@dataclass
class BenchmarkData:
    """Structured benchmark data"""
    protocol: str
    mode: str
    work_factor: int
    payload_size_bytes: int
    concurrency: int
    duration_seconds: float
    rps: float
    p50: float
    p90: float
    p95: float
    p99: float
    mean: float
    min_latency: float
    max_latency: float
    error_rate: float
    total_requests: int
    successful_requests: int
    failed_requests: int
    timestamp: str

    @classmethod
    def from_file(cls, filepath: str):
        """Load benchmark data from JSON file"""
        with open(filepath, 'r') as f:
            data = json.load(f)
        
        stats = data['stats']
        return cls(
            protocol=data['protocol'],
            mode=data['mode'],
            work_factor=data['work_factor'],
            payload_size_bytes=data['payload_size_bytes'],
            concurrency=data['concurrency'],
            duration_seconds=data['duration'],
            rps=stats['rps'],
            p50=stats['p50'],
            p90=stats['p90'],
            p95=stats['p95'],
            p99=stats['p99'],
            mean=stats['mean'],
            min_latency=stats['min'],
            max_latency=stats['max'],
            error_rate=stats['error_rate'],
            total_requests=stats['total_requests'],
            successful_requests=stats['successful_requests'],
            failed_requests=stats['failed_requests'],
            timestamp=data['timestamp']
        )


def load_results(results_dir: str) -> List[BenchmarkData]:
    """Load all benchmark results from directory"""
    pattern = os.path.join(results_dir, "*.json")
    files = glob.glob(pattern)
    
    if not files:
        print(f"No result files found in {results_dir}")
        return []
    
    results = []
    for filepath in files:
        try:
            data = BenchmarkData.from_file(filepath)
            results.append(data)
        except Exception as e:
            print(f"Warning: Failed to load {filepath}: {e}")
    
    print(f"Loaded {len(results)} benchmark results")
    return results


def group_by_mode_and_payload(results: List[BenchmarkData]) -> Dict:
    """Group results by mode and payload size"""
    grouped = {}
    for result in results:
        key = (result.mode, result.payload_size_bytes, result.work_factor)
        if key not in grouped:
            grouped[key] = []
        grouped[key].append(result)
    return grouped


def generate_charts(results: List[BenchmarkData], output_dir: str):
    """Generate comparison charts"""
    os.makedirs(output_dir, exist_ok=True)
    
    grouped = group_by_mode_and_payload(results)
    
    for (mode, payload, work_factor), data_list in grouped.items():
        # Sort by concurrency
        data_list.sort(key=lambda x: x.concurrency)
        
        # Group by protocol
        protocols = {}
        for d in data_list:
            if d.protocol not in protocols:
                protocols[d.protocol] = []
            protocols[d.protocol].append(d)
        
        # Create figure with subplots
        fig, ((ax1, ax2), (ax3, ax4)) = plt.subplots(2, 2, figsize=(16, 12))
        fig.suptitle(f'Protocol Comparison: mode={mode}, payload={payload}B, work_factor={work_factor}', 
                     fontsize=14, fontweight='bold')
        
        # Chart 1: Throughput vs Concurrency
        for protocol, pdata in protocols.items():
            concurrency = [d.concurrency for d in pdata]
            rps = [d.rps for d in pdata]
            ax1.plot(concurrency, rps, marker='o', label=protocol.upper(), linewidth=2)
        ax1.set_xlabel('Concurrency', fontweight='bold')
        ax1.set_ylabel('Throughput (RPS)', fontweight='bold')
        ax1.set_title('Throughput vs Concurrency')
        ax1.legend()
        ax1.grid(True, alpha=0.3)
        
        # Chart 2: p95 Latency vs Concurrency
        for protocol, pdata in protocols.items():
            concurrency = [d.concurrency for d in pdata]
            p95 = [d.p95 for d in pdata]
            ax2.plot(concurrency, p95, marker='s', label=protocol.upper(), linewidth=2)
        ax2.set_xlabel('Concurrency', fontweight='bold')
        ax2.set_ylabel('p95 Latency (ms)', fontweight='bold')
        ax2.set_title('p95 Latency vs Concurrency')
        ax2.legend()
        ax2.grid(True, alpha=0.3)
        
        # Chart 3: p99 Latency vs Concurrency
        for protocol, pdata in protocols.items():
            concurrency = [d.concurrency for d in pdata]
            p99 = [d.p99 for d in pdata]
            ax3.plot(concurrency, p99, marker='^', label=protocol.upper(), linewidth=2)
        ax3.set_xlabel('Concurrency', fontweight='bold')
        ax3.set_ylabel('p99 Latency (ms)', fontweight='bold')
        ax3.set_title('p99 Latency vs Concurrency')
        ax3.legend()
        ax3.grid(True, alpha=0.3)
        
        # Chart 4: Error Rate vs Concurrency
        for protocol, pdata in protocols.items():
            concurrency = [d.concurrency for d in pdata]
            error_rate = [d.error_rate * 100 for d in pdata]
            ax4.plot(concurrency, error_rate, marker='d', label=protocol.upper(), linewidth=2)
        ax4.set_xlabel('Concurrency', fontweight='bold')
        ax4.set_ylabel('Error Rate (%)', fontweight='bold')
        ax4.set_title('Error Rate vs Concurrency')
        ax4.legend()
        ax4.grid(True, alpha=0.3)
        
        # Save figure
        filename = f"{mode}_payload{payload}B_wf{work_factor}.png"
        filepath = os.path.join(output_dir, filename)
        plt.tight_layout()
        plt.savefig(filepath, dpi=150, bbox_inches='tight')
        plt.close()
        print(f"Generated: {filepath}")


def generate_summary(results: List[BenchmarkData]) -> str:
    """Generate text summary of results"""
    if not results:
        return "No results to summarize"
    
    summary = []
    summary.append("# Benchmark Summary\n")
    
    grouped = group_by_mode_and_payload(results)
    
    for (mode, payload, work_factor), data_list in grouped.items():
        summary.append(f"\n## {mode.upper()} Mode - Payload: {payload}B - Work Factor: {work_factor}\n")
        
        # Group by protocol
        protocols = {}
        for d in data_list:
            if d.protocol not in protocols:
                protocols[d.protocol] = []
            protocols[d.protocol].append(d)
        
        for protocol, pdata in sorted(protocols.items()):
            summary.append(f"\n### {protocol.upper()}\n")
            summary.append("| Concurrency | RPS | p50 (ms) | p95 (ms) | p99 (ms) | Mean (ms) | Error Rate |")
            summary.append("|------------|-----|----------|----------|----------|-----------|------------|")
            
            for d in sorted(pdata, key=lambda x: x.concurrency):
                summary.append(f"| {d.concurrency:10} | {d.rps:7.1f} | {d.p50:8.2f} | {d.p95:8.2f} | {d.p99:8.2f} | {d.mean:9.2f} | {d.error_rate*100:9.2f}% |")
    
    return "\n".join(summary)


def main():
    if len(sys.argv) < 2:
        print("Usage: python analyze.py <results_directory>")
        print("Example: python analyze.py ../results")
        sys.exit(1)
    
    results_dir = sys.argv[1]
    
    if not os.path.isdir(results_dir):
        print(f"Error: {results_dir} is not a directory")
        sys.exit(1)
    
    # Load results
    results = load_results(results_dir)
    
    if not results:
        print("No valid results found")
        sys.exit(1)
    
    # Generate charts
    charts_dir = os.path.join(results_dir, "charts")
    print(f"\nGenerating charts in {charts_dir}...")
    generate_charts(results, charts_dir)
    
    # Generate summary
    print("\nGenerating summary...")
    summary = generate_summary(results)
    
    summary_file = os.path.join(results_dir, "BENCHMARK_SUMMARY.md")
    with open(summary_file, 'w') as f:
        f.write(summary)
    
    print(f"Summary saved to: {summary_file}")
    print("\n✓ Analysis complete")


if __name__ == "__main__":
    main()
