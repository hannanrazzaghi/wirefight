package results

import (
	"encoding/json"
	"math"
	"os"
	"sort"
	"time"
)

// RequestResult captures the result of a single request
type RequestResult struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	LatencyMs float64   `json:"latency_ms"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

// AggregatedStats contains aggregated statistics from a load test run
type AggregatedStats struct {
	TotalRequests      int     `json:"total_requests"`
	SuccessfulRequests int     `json:"successful_requests"`
	FailedRequests     int     `json:"failed_requests"`
	RPS                float64 `json:"rps"`
	P50                float64 `json:"p50"`
	P90                float64 `json:"p90"`
	P95                float64 `json:"p95"`
	P99                float64 `json:"p99"`
	Mean               float64 `json:"mean"`
	Min                float64 `json:"min"`
	Max                float64 `json:"max"`
	ErrorRate          float64 `json:"error_rate"`
}

// BenchmarkResult contains the full benchmark result
type BenchmarkResult struct {
	Protocol         string          `json:"protocol"`
	Mode             string          `json:"mode"`
	WorkFactor       int             `json:"work_factor"`
	PayloadSizeBytes int             `json:"payload_size_bytes"`
	Concurrency      int             `json:"concurrency"`
	Duration         float64         `json:"duration_seconds"`
	WarmupDuration   float64         `json:"warmup_duration_seconds"`
	Stats            AggregatedStats `json:"stats"`
	Timestamp        string          `json:"timestamp"`
}

// Collector collects and aggregates request results
type Collector struct {
	results []RequestResult
}

// NewCollector creates a new results collector
func NewCollector() *Collector {
	return &Collector{
		results: make([]RequestResult, 0, 10000),
	}
}

// Add adds a request result
func (c *Collector) Add(result RequestResult) {
	c.results = append(c.results, result)
}

// Aggregate computes aggregated statistics from collected results
func (c *Collector) Aggregate(duration time.Duration) AggregatedStats {
	if len(c.results) == 0 {
		return AggregatedStats{}
	}

	// Separate successful and failed requests
	var successLatencies []float64
	successCount := 0
	failCount := 0

	for _, r := range c.results {
		if r.Success {
			successLatencies = append(successLatencies, r.LatencyMs)
			successCount++
		} else {
			failCount++
		}
	}

	stats := AggregatedStats{
		TotalRequests:      len(c.results),
		SuccessfulRequests: successCount,
		FailedRequests:     failCount,
		RPS:                float64(successCount) / duration.Seconds(),
		ErrorRate:          float64(failCount) / float64(len(c.results)),
	}

	if len(successLatencies) == 0 {
		return stats
	}

	// Sort latencies for percentile calculation
	sort.Float64s(successLatencies)

	stats.P50 = percentile(successLatencies, 50)
	stats.P90 = percentile(successLatencies, 90)
	stats.P95 = percentile(successLatencies, 95)
	stats.P99 = percentile(successLatencies, 99)
	stats.Min = successLatencies[0]
	stats.Max = successLatencies[len(successLatencies)-1]
	stats.Mean = mean(successLatencies)

	return stats
}

// percentile calculates the given percentile from sorted data
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	index := (p / 100.0) * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// mean calculates the arithmetic mean
func mean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// SaveToFile saves the benchmark result to a JSON file
func SaveToFile(result BenchmarkResult, filename string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
