package logic

import (
	"testing"
	"time"
)

func TestExecute_CPUMode(t *testing.T) {
	req := Request{
		RequestID:        "test-1",
		Mode:             ModeCPU,
		WorkFactor:       100,
		PayloadSizeBytes: 256,
	}

	resp, err := Execute(req, "test")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if resp.RequestID != req.RequestID {
		t.Errorf("RequestID mismatch: got %s, want %s", resp.RequestID, req.RequestID)
	}
	if resp.Mode != ModeCPU {
		t.Errorf("Mode mismatch: got %s, want %s", resp.Mode, ModeCPU)
	}
	if resp.Result == "" {
		t.Error("Result should not be empty")
	}
	if resp.ServerProcessingMs <= 0 {
		t.Error("ServerProcessingMs should be positive")
	}
	if resp.Protocol != "test" {
		t.Errorf("Protocol mismatch: got %s, want test", resp.Protocol)
	}
}

func TestExecute_IOMode(t *testing.T) {
	req := Request{
		RequestID:        "test-2",
		Mode:             ModeIO,
		WorkFactor:       50,
		PayloadSizeBytes: 1024,
	}

	start := time.Now()
	resp, err := Execute(req, "test")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// IO mode should sleep for approximately work_factor ms
	if elapsed < 40*time.Millisecond {
		t.Errorf("IO work completed too quickly: %v", elapsed)
	}

	if resp.Result != "slept_50ms" {
		t.Errorf("Result mismatch: got %s, want slept_50ms", resp.Result)
	}
}

func TestExecute_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     Request
		wantErr bool
	}{
		{
			name: "missing request_id",
			req: Request{
				Mode:       ModeCPU,
				WorkFactor: 10,
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			req: Request{
				RequestID:  "test",
				Mode:       "invalid",
				WorkFactor: 10,
			},
			wantErr: true,
		},
		{
			name: "negative work_factor",
			req: Request{
				RequestID:  "test",
				Mode:       ModeCPU,
				WorkFactor: -1,
			},
			wantErr: true,
		},
		{
			name: "negative payload_size",
			req: Request{
				RequestID:        "test",
				Mode:             ModeCPU,
				WorkFactor:       10,
				PayloadSizeBytes: -1,
			},
			wantErr: true,
		},
		{
			name: "valid zero work_factor",
			req: Request{
				RequestID:  "test",
				Mode:       ModeCPU,
				WorkFactor: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Execute(tt.req, "test")
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecuteCPUWork_Deterministic(t *testing.T) {
	// Same input should produce same result
	result1 := executeCPUWork(100, 0)
	result2 := executeCPUWork(100, 0)

	if result1 != result2 {
		t.Errorf("CPU work not deterministic: %s != %s", result1, result2)
	}
}

func BenchmarkExecute_CPU_Light(b *testing.B) {
	req := Request{
		RequestID:        "bench",
		Mode:             ModeCPU,
		WorkFactor:       10,
		PayloadSizeBytes: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Execute(req, "bench")
	}
}

func BenchmarkExecute_CPU_Heavy(b *testing.B) {
	req := Request{
		RequestID:        "bench",
		Mode:             ModeCPU,
		WorkFactor:       1000,
		PayloadSizeBytes: 4096,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Execute(req, "bench")
	}
}
