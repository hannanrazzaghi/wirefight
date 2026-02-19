package results

import (
	"testing"
	"time"
)

func TestPercentile(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	tests := []struct {
		p    float64
		want float64
	}{
		{50, 5.5},
		{90, 9.1},
		{95, 9.55},
		{99, 9.91},
	}

	for _, tt := range tests {
		got := percentile(data, tt.p)
		if got < tt.want-0.1 || got > tt.want+0.1 {
			t.Errorf("percentile(%v, %.0f) = %.2f, want ~%.2f", data, tt.p, got, tt.want)
		}
	}
}

func TestMean(t *testing.T) {
	tests := []struct {
		data []float64
		want float64
	}{
		{[]float64{1, 2, 3, 4, 5}, 3.0},
		{[]float64{10, 20, 30}, 20.0},
		{[]float64{5}, 5.0},
		{[]float64{}, 0.0},
	}

	for _, tt := range tests {
		got := mean(tt.data)
		if got != tt.want {
			t.Errorf("mean(%v) = %.2f, want %.2f", tt.data, got, tt.want)
		}
	}
}

func TestCollector_Aggregate(t *testing.T) {
	collector := NewCollector()

	// Add successful requests
	for i := 0; i < 100; i++ {
		collector.Add(RequestResult{
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Duration(i+1) * time.Millisecond),
			LatencyMs: float64(i + 1),
			Success:   true,
		})
	}

	// Add failed requests
	for i := 0; i < 10; i++ {
		collector.Add(RequestResult{
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Success:   false,
			Error:     "test error",
		})
	}

	stats := collector.Aggregate(10 * time.Second)

	if stats.TotalRequests != 110 {
		t.Errorf("TotalRequests = %d, want 110", stats.TotalRequests)
	}

	if stats.SuccessfulRequests != 100 {
		t.Errorf("SuccessfulRequests = %d, want 100", stats.SuccessfulRequests)
	}

	if stats.FailedRequests != 10 {
		t.Errorf("FailedRequests = %d, want 10", stats.FailedRequests)
	}

	if stats.RPS != 10.0 {
		t.Errorf("RPS = %.2f, want 10.00", stats.RPS)
	}

	if stats.Min != 1.0 {
		t.Errorf("Min = %.2f, want 1.00", stats.Min)
	}

	if stats.Max != 100.0 {
		t.Errorf("Max = %.2f, want 100.00", stats.Max)
	}

	if stats.Mean != 50.5 {
		t.Errorf("Mean = %.2f, want 50.50", stats.Mean)
	}

	errorRate := float64(10) / float64(110)
	if stats.ErrorRate < errorRate-0.01 || stats.ErrorRate > errorRate+0.01 {
		t.Errorf("ErrorRate = %.4f, want ~%.4f", stats.ErrorRate, errorRate)
	}
}

func TestCollector_EmptyResults(t *testing.T) {
	collector := NewCollector()
	stats := collector.Aggregate(1 * time.Second)

	if stats.TotalRequests != 0 {
		t.Errorf("Expected 0 total requests, got %d", stats.TotalRequests)
	}

	if stats.RPS != 0 {
		t.Errorf("Expected 0 RPS, got %.2f", stats.RPS)
	}
}

func TestCollector_OnlyFailures(t *testing.T) {
	collector := NewCollector()

	for i := 0; i < 5; i++ {
		collector.Add(RequestResult{
			Success: false,
			Error:   "error",
		})
	}

	stats := collector.Aggregate(1 * time.Second)

	if stats.TotalRequests != 5 {
		t.Errorf("TotalRequests = %d, want 5", stats.TotalRequests)
	}

	if stats.FailedRequests != 5 {
		t.Errorf("FailedRequests = %d, want 5", stats.FailedRequests)
	}

	if stats.ErrorRate != 1.0 {
		t.Errorf("ErrorRate = %.2f, want 1.00", stats.ErrorRate)
	}

	// All stats should be zero when there are no successful requests
	if stats.P50 != 0 || stats.P90 != 0 || stats.Mean != 0 {
		t.Error("Expected all latency stats to be 0 with only failures")
	}
}
