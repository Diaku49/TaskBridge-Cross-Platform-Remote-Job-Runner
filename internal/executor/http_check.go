package executor

import (
	"context"
	"fmt"
	"net/http"
	"taskbridge/internal/model"
)

type HTTPCheckExecutor struct {
	Client *http.Client
}

func NewHTTPCheckExecutor(client *http.Client) *HTTPCheckExecutor {
	return &HTTPCheckExecutor{Client: client}
}

type HTTPCheckPayload struct {
	URL            string `json:"url"`
	Method         string `json:"method,omitempty"`
	ExpectedStatus int    `json:"expected_status"`
}

func (e *HTTPCheckExecutor) Type() model.JobType {
	return model.JobHTTPCheck
}

func (e *HTTPCheckExecutor) Execute(ctx context.Context, job model.Job) Result {
	logs := make([]string, 0, 4)

	var payload HTTPCheckPayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "invalid payload",
			Logs:   []string{"failed to decode payload: " + err.Error()},
			Result: nil,
		}
	}
	logs = append(logs, "decoded payload successfully")

	if err := CheckPayload(&payload); err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "invalid payload",
			Logs:   append(logs, "payload validation failed: "+err.Error()),
			Result: nil,
		}
	}

	req, err := http.NewRequestWithContext(ctx, payload.Method, payload.URL, nil)
	if err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "failed to create HTTP request",
			Logs:   append(logs, "failed to create HTTP request: "+err.Error()),
			Result: nil,
		}
	}
	logs = append(logs, fmt.Sprintf("created HTTP request: method=%s, url=%s", payload.Method, payload.URL))

	resp, err := e.Client.Do(req)
	if err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "HTTP request failed",
			Logs:   append(logs, "HTTP request failed: "+err.Error()),
			Result: nil,
		}
	}
	defer resp.Body.Close()
	logs = append(logs, fmt.Sprintf("received HTTP response: status=%d", resp.StatusCode))

	if resp.StatusCode != payload.ExpectedStatus {
		return Result{
			Status: model.JobFailed,
			Error:  "unexpected status code",
			Logs:   append(logs, fmt.Sprintf("expected status: %d, but got: %d", payload.ExpectedStatus, resp.StatusCode)),
			Result: map[string]any{
				"received_status": resp.StatusCode,
				"expected_status": payload.ExpectedStatus,
				"url":             payload.URL,
				"method":          payload.Method,
			},
		}
	}

	return Result{
		Status: model.JobSuccess,
		Logs:   logs,
		Result: map[string]any{
			"received_status": resp.StatusCode,
			"expected_status": payload.ExpectedStatus,
			"url":             payload.URL,
			"method":          payload.Method,
		},
	}
}

func CheckPayload(p *HTTPCheckPayload) error {
	if p.URL == "" {
		return fmt.Errorf("url is required")
	}

	if p.Method == "" {
		p.Method = http.MethodGet
	}

	if p.ExpectedStatus == 0 {
		return fmt.Errorf("expected_status is required")
	}

	return nil
}
