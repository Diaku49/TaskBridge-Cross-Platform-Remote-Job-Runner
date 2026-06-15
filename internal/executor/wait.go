package executor

import (
	"context"
	"fmt"
	"taskbridge/internal/model"
	"time"
)

type WaitExecutor struct{}

type WaitPayload struct {
	DurationSeconds int `json:"duration_seconds"`
}

func NewWaitExecutor() *WaitExecutor {
	return &WaitExecutor{}
}

func (e *WaitExecutor) Type() model.JobType {
	return model.JobWait
}

func (e *WaitExecutor) Execute(ctx context.Context, job model.Job) Result {
	logs := make([]string, 0, 4)

	var payload WaitPayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "invalid payload",
			Logs:   []string{"failed to decode payload: " + err.Error()},
		}
	}

	logs = append(logs, "decoded payload successfully")

	if err := CheckWaitPayload(&payload); err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "invalid payload",
			Logs:   append(logs, "payload validation failed: "+err.Error()),
		}
	}

	duration := time.Duration(payload.DurationSeconds) * time.Second
	logs = append(logs, fmt.Sprintf("waiting for %d seconds", payload.DurationSeconds))

	timer := time.NewTimer(duration)
	defer timer.Stop()

	start := time.Now().UTC()

	select {
	case <-ctx.Done():
		return Result{
			Status: model.JobFailed,
			Error:  ctx.Err().Error(),
			Logs:   append(logs, "wait job canceled or timed out"),
			Result: map[string]any{
				"duration_seconds": payload.DurationSeconds,
				"completed":        false,
				"elapsed_seconds":  int(time.Since(start).Seconds()),
			},
		}

	case <-timer.C:
		logs = append(logs, "wait completed successfully")

		return Result{
			Status: model.JobSuccess,
			Logs:   logs,
			Result: map[string]any{
				"duration_seconds": payload.DurationSeconds,
				"completed":        true,
				"elapsed_seconds":  int(time.Since(start).Seconds()),
			},
		}
	}
}

func CheckWaitPayload(p *WaitPayload) error {
	if p.DurationSeconds <= 0 {
		return fmt.Errorf("duration_seconds must be greater than 0")
	}

	return nil
}
