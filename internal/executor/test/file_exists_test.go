package executor_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"taskbridge/internal/executor"
	"taskbridge/internal/model"
)

func TestFileExistsExecutorExecuteExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	ex := executor.NewFileExistsExecutor()
	result := ex.Execute(context.Background(), model.Job{
		Type: model.JobFileExists,
		Payload: map[string]any{
			"path": path,
		},
	})

	if result.Status != model.JobSuccess {
		t.Fatalf("expected status %s, got %s: %s", model.JobSuccess, result.Status, result.Error)
	}
	if result.Result["exists"] != true {
		t.Fatalf("expected exists true, got %v", result.Result["exists"])
	}
}

func TestFileExistsExecutorExecuteMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.txt")

	ex := executor.NewFileExistsExecutor()
	result := ex.Execute(context.Background(), model.Job{
		Type: model.JobFileExists,
		Payload: map[string]any{
			"path": path,
		},
	})

	if result.Status != model.JobFailed {
		t.Fatalf("expected status %s, got %s", model.JobFailed, result.Status)
	}
	if result.Error != "file does not exist" {
		t.Fatalf("expected file does not exist error, got %q", result.Error)
	}
	if result.Result["exists"] != false {
		t.Fatalf("expected exists false, got %v", result.Result["exists"])
	}
}
