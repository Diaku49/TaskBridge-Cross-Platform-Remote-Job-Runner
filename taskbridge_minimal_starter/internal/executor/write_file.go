package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"taskbridge/internal/model"
)

type WriteFileExecutor struct{}

type WriteFilePayload struct {
	Path       string `json:"path"`
	Content    string `json:"content"`
	Append     bool   `json:"append,omitempty"`
	CreateDirs bool   `json:"create_dirs,omitempty"`
}

func NewWriteFileExecutor() *WriteFileExecutor {
	return &WriteFileExecutor{}
}

func (e *WriteFileExecutor) Type() model.JobType {
	return model.JobWriteFile
}

func (e *WriteFileExecutor) Execute(ctx context.Context, job model.Job) Result {
	logs := make([]string, 0, 5)

	var payload WriteFilePayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "invalid payload",
			Logs:   []string{"failed to decode payload: " + err.Error()},
		}
	}

	logs = append(logs, "decoded payload successfully")

	if err := CheckWriteFilePayload(&payload); err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "invalid payload",
			Logs:   append(logs, "payload validation failed: "+err.Error()),
		}
	}

	if err := ctx.Err(); err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  err.Error(),
			Logs:   append(logs, "write file job canceled or timed out before writing"),
			Result: map[string]any{
				"path": payload.Path,
			},
		}
	}

	if payload.CreateDirs {
		dir := filepath.Dir(payload.Path)

		if err := os.MkdirAll(dir, 0755); err != nil {
			return Result{
				Status: model.JobFailed,
				Error:  "failed to create parent directories",
				Logs:   append(logs, "failed to create parent directories: "+err.Error()),
				Result: map[string]any{
					"path": payload.Path,
				},
			}
		}

		logs = append(logs, "parent directories checked/created successfully")
	}

	var err error
	var bytesWritten int

	if payload.Append {
		bytesWritten, err = appendToFile(payload.Path, payload.Content)
	} else {
		bytesWritten, err = overwriteFile(payload.Path, payload.Content)
	}

	if err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "failed to write file",
			Logs:   append(logs, "failed to write file: "+err.Error()),
			Result: map[string]any{
				"path":   payload.Path,
				"append": payload.Append,
			},
		}
	}

	if payload.Append {
		logs = append(logs, "content appended to file successfully")
	} else {
		logs = append(logs, "file written successfully")
	}

	return Result{
		Status: model.JobSuccess,
		Logs:   logs,
		Result: map[string]any{
			"path":          payload.Path,
			"append":        payload.Append,
			"bytes_written": bytesWritten,
		},
	}
}

func CheckWriteFilePayload(p *WriteFilePayload) error {
	if p.Path == "" {
		return fmt.Errorf("path is required")
	}

	return nil
}

func overwriteFile(path string, content string) (int, error) {
	data := []byte(content)

	if err := os.WriteFile(path, data, 0644); err != nil {
		return 0, err
	}

	return len(data), nil
}

func appendToFile(path string, content string) (int, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return file.WriteString(content)
}
