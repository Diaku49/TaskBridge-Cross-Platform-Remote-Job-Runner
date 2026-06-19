package executor_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"taskbridge/internal/executor"
	"taskbridge/internal/model"
)

func TestChecksumExecutorExecuteSHA256Success(t *testing.T) {
	content := []byte("hello checksum")
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	expectedSum := sha256.Sum256(content)
	ex := executor.NewChecksumExecutor()
	result := ex.Execute(context.Background(), model.Job{
		Type: model.JobChecksum,
		Payload: map[string]any{
			"path":      path,
			"algorithm": "sha256",
		},
	})

	if result.Status != model.JobSuccess {
		t.Fatalf("expected status %s, got %s: %s", model.JobSuccess, result.Status, result.Error)
	}
	if result.Result["checksum"] != hex.EncodeToString(expectedSum[:]) {
		t.Fatalf("unexpected checksum: %v", result.Result["checksum"])
	}
	if result.Result["bytes_read"] != int64(len(content)) {
		t.Fatalf("expected bytes_read %d, got %v", len(content), result.Result["bytes_read"])
	}
}

func TestChecksumExecutorExecuteUnsupportedAlgorithm(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	ex := executor.NewChecksumExecutor()
	result := ex.Execute(context.Background(), model.Job{
		Type: model.JobChecksum,
		Payload: map[string]any{
			"path":      path,
			"algorithm": "crc32",
		},
	})

	if result.Status != model.JobFailed {
		t.Fatalf("expected status %s, got %s", model.JobFailed, result.Status)
	}
	if result.Error != "invalid payload" {
		t.Fatalf("expected invalid payload error, got %q", result.Error)
	}
}
