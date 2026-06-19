package executor

import (
	"context"
	"fmt"
	"os"
	"taskbridge/internal/model"
)

type FileExistsExecutor struct{}

type FileExistsPayload struct {
	Path string `json:"path"`
}

func NewFileExistsExecutor() *FileExistsExecutor {
	return &FileExistsExecutor{}
}

func (e *FileExistsExecutor) Type() model.JobType {
	return model.JobFileExists
}

func (e *FileExistsExecutor) Execute(ctx context.Context, job model.Job) Result {
	logs := make([]string, 0, 4)

	var payload FileExistsPayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "invalid payload",
			Logs:   []string{"failed to decode payload: " + err.Error()},
		}
	}

	logs = append(logs, "decoded payload successfully")

	if err := CheckFileExistsPayload(&payload); err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "invalid payload",
			Logs:   append(logs, "payload validation failed: "+err.Error()),
		}
	}

	// Check cancellation or timeout
	select {
	case <-ctx.Done():
		return Result{
			Status: model.JobFailed,
			Error:  ctx.Err().Error(),
			Logs:   append(logs, "file existence check canceled or timed out"),
			Result: map[string]any{
				"path":   payload.Path,
				"exists": false,
			},
		}
	default:
	}

	info, err := os.Stat(payload.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return Result{
				Status: model.JobFailed,
				Error:  "file does not exist",
				Logs:   append(logs, "path does not exist"),
				Result: map[string]any{
					"path":   payload.Path,
					"exists": false,
				},
			}
		}

		return Result{
			Status: model.JobFailed,
			Error:  "failed to check file",
			Logs:   append(logs, "failed to stat path: "+err.Error()),
			Result: map[string]any{
				"path":   payload.Path,
				"exists": false,
			},
		}
	}

	logs = append(logs, "path exists")

	return Result{
		Status: model.JobSuccess,
		Logs:   logs,
		Result: map[string]any{
			"path":     payload.Path,
			"exists":   true,
			"is_dir":   info.IsDir(),
			"size":     info.Size(),
			"mod_time": info.ModTime().UTC(),
		},
	}
}

func CheckFileExistsPayload(p *FileExistsPayload) error {
	if p.Path == "" {
		return fmt.Errorf("path is required")
	}

	return nil
}
