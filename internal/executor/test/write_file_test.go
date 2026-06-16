package executor_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"taskbridge/internal/executor"
	"taskbridge/internal/model"
)

func TestWriteFileExecutorExecuteCreatesParentDirs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "output.txt")

	ex := executor.NewWriteFileExecutor()
	result := ex.Execute(context.Background(), model.Job{
		Type: model.JobWriteFile,
		Payload: map[string]any{
			"path":        path,
			"content":     "hello writer",
			"create_dirs": true,
		},
	})

	if result.Status != model.JobSuccess {
		t.Fatalf("expected status %s, got %s: %s", model.JobSuccess, result.Status, result.Error)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(content) != "hello writer" {
		t.Fatalf("unexpected file content: %q", string(content))
	}
}

func TestWriteFileExecutorExecuteAppendsContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output.txt")
	if err := os.WriteFile(path, []byte("first\n"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	ex := executor.NewWriteFileExecutor()
	result := ex.Execute(context.Background(), model.Job{
		Type: model.JobWriteFile,
		Payload: map[string]any{
			"path":    path,
			"content": "second\n",
			"append":  true,
		},
	})

	if result.Status != model.JobSuccess {
		t.Fatalf("expected status %s, got %s: %s", model.JobSuccess, result.Status, result.Error)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(content) != "first\nsecond\n" {
		t.Fatalf("unexpected file content: %q", string(content))
	}
}
