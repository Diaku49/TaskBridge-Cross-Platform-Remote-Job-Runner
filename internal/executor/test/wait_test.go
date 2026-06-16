package executor_test

import (
	"context"
	"testing"
	"time"

	"taskbridge/internal/executor"
	"taskbridge/internal/model"
)

func TestWaitExecutorExecuteCompletes(t *testing.T) {
	ex := executor.NewWaitExecutor()
	start := time.Now()
	result := ex.Execute(context.Background(), model.Job{
		Type: model.JobWait,
		Payload: map[string]any{
			"duration_seconds": 1,
		},
	})

	if result.Status != model.JobSuccess {
		t.Fatalf("expected status %s, got %s: %s", model.JobSuccess, result.Status, result.Error)
	}
	if time.Since(start) < time.Second {
		t.Fatalf("wait completed too early")
	}
	if result.Result["completed"] != true {
		t.Fatalf("expected completed true, got %v", result.Result["completed"])
	}
}

func TestWaitExecutorExecuteCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ex := executor.NewWaitExecutor()
	result := ex.Execute(ctx, model.Job{
		Type: model.JobWait,
		Payload: map[string]any{
			"duration_seconds": 1,
		},
	})

	if result.Status != model.JobFailed {
		t.Fatalf("expected status %s, got %s", model.JobFailed, result.Status)
	}
	if result.Error != context.Canceled.Error() {
		t.Fatalf("expected canceled error, got %q", result.Error)
	}
	if result.Result["completed"] != false {
		t.Fatalf("expected completed false, got %v", result.Result["completed"])
	}
}
