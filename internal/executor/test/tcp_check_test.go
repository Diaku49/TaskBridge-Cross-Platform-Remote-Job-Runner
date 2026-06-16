package executor_test

import (
	"context"
	"net"
	"testing"

	"taskbridge/internal/executor"
	"taskbridge/internal/model"
)

func TestTCPCheckExecutorExecuteSuccess(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err == nil {
			_ = conn.Close()
		}
	}()

	ex := executor.NewTcpCheckExecutor()
	result := ex.Execute(context.Background(), model.Job{
		Type: model.JobTCPCheck,
		Payload: map[string]any{
			"address": listener.Addr().String(),
		},
	})

	if result.Status != model.JobSuccess {
		t.Fatalf("expected status %s, got %s: %s", model.JobSuccess, result.Status, result.Error)
	}
	if result.Result["reachable"] != true {
		t.Fatalf("expected reachable true, got %v", result.Result["reachable"])
	}
}

func TestTCPCheckExecutorExecuteMissingAddress(t *testing.T) {
	ex := executor.NewTcpCheckExecutor()
	result := ex.Execute(context.Background(), model.Job{
		Type:    model.JobTCPCheck,
		Payload: map[string]any{},
	})

	if result.Status != model.JobFailed {
		t.Fatalf("expected status %s, got %s", model.JobFailed, result.Status)
	}
	if result.Error != "invalid payload" {
		t.Fatalf("expected invalid payload error, got %q", result.Error)
	}
}
