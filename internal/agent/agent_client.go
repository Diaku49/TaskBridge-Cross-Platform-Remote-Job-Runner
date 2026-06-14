package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	e "taskbridge/internal/executor"
	"taskbridge/internal/model"
	"time"
)

type AgentClient struct {
	ServerURL        string
	PollInterval     time.Duration
	ExecutorRegistry *e.Registry
	Agent            *model.Agent
	HTTPClient       *http.Client
}

var validJobTypes = map[model.JobType]struct{}{
	model.JobHTTPCheck:  {},
	model.JobTCPCheck:   {},
	model.JobFileExists: {},
	model.JobChecksum:   {},
	model.JobWriteFile:  {},
	model.JobWait:       {},
}

func NewAgentClient(agentId, serverURL, capabilities string, pollInterval time.Duration) *AgentClient {
	caps := parseCapabilities(capabilities)
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	osName := runtime.GOOS
	arch := runtime.GOARCH
	execRegistery := NewRegistry(caps)

	return &AgentClient{
		ServerURL:        serverURL,
		PollInterval:     pollInterval,
		ExecutorRegistry: execRegistery,

		Agent: &model.Agent{
			ID:           agentId,
			Capabilities: caps,
			Hostname:     hostname,
			OS:           osName,
			Arch:         arch,
			Version:      "1.0",
		},
		HTTPClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (ac *AgentClient) Register() error {
	registerInfo := model.RegisterAgentRequest{
		ID:           ac.Agent.ID,
		Hostname:     ac.Agent.Hostname,
		OS:           ac.Agent.OS,
		Arch:         ac.Agent.Arch,
		Version:      ac.Agent.Version,
		Capabilities: ac.Agent.Capabilities,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(registerInfo); err != nil {
		return fmt.Errorf(" failed to encode register request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, ac.ServerURL+"/agents/register", &buf)
	if err != nil {
		return fmt.Errorf(" failed to create register request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Registration
	resp, err := ac.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf(" failed to send register request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp model.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("registration failed with status: %s, and failed to parse error response: %v", resp.Status, err)
		}

		return fmt.Errorf("registration failed with status: %s, error: %s", resp.Status, errResp.Message)
	}

	return nil
}

func (ac *AgentClient) Start(ctx context.Context) {
	go HeartbeatLoop(ac, ctx)

	ticker := time.NewTicker(ac.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Agent client stopped")
			return
		case <-ticker.C:
			job, hasJob, err := ac.Poll()
			if err != nil {
				fmt.Printf("Error polling for job: %v\n", err)
				continue
			}
			if !hasJob {
				fmt.Println("No job available, will poll again...")
				continue
			}

			ac.ExecuteJob(job, ctx)
		}
	}
}

func (ac *AgentClient) ExecuteJob(job model.Job, ctx context.Context) {
	// find Job
	executor, found := ac.ExecutorRegistry.Get(job.Type)
	if !found {
		result := e.Result{
			Status: model.JobFailed,
			Error:  fmt.Sprintf("unsupported job type: %s", job.Type),
			Logs:   []string{fmt.Sprintf("no executor found for this job type: %s", job.Type)},
			Result: nil,
		}

		if err := ac.SubmitJobResult(job.ID, result); err != nil {
			fmt.Printf("Failed to submit unsupported-job result: %v\n", err)
		}

		return
	}

	// Checking timeout
	jobCtx := ctx
	var cancel context.CancelFunc
	if job.TimeoutSeconds > 0 {
		jobCtx, cancel = context.WithTimeout(ctx, time.Duration(job.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	result := executor.Execute(jobCtx, job)

	if err := ac.SubmitJobResult(job.ID, result); err != nil {
		fmt.Printf("Failed to submit job result: %v\n", err)
	}
}

func (ac *AgentClient) SubmitJobResult(jobId string, result e.Result) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(result); err != nil {
		return fmt.Errorf("failed to encode job result: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/jobs/%s/result", ac.ServerURL, jobId), &buf)
	if err != nil {
		return fmt.Errorf("failed to create submit job result request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send submit job result request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp model.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("submit job result failed with status: %s, and failed to parse error response: %v", resp.Status, err)
		}

		return fmt.Errorf("submit job result failed with status: %s, error: %s", resp.Status, errResp.Message)
	}

	return nil
}

func (ac *AgentClient) Poll() (model.Job, bool, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/agents/%s/next-job", ac.ServerURL, ac.Agent.ID), nil)
	if err != nil {
		return model.Job{}, false, fmt.Errorf("failed to create next job request: %w", err)
	}

	resp, err := ac.HTTPClient.Do(req)
	if err != nil {
		return model.Job{}, false, fmt.Errorf("failed to send next job request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp model.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return model.Job{}, false, fmt.Errorf("poll failed with status: %s, and failed to parse error response: %v", resp.Status, err)
		}

		return model.Job{}, false, fmt.Errorf("poll failed with status: %s, error: %s", resp.Status, errResp.Message)
	}

	if resp.StatusCode == http.StatusNoContent {
		return model.Job{}, false, nil
	}

	var job model.Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return model.Job{}, false, fmt.Errorf("failed to decode next job response: %w", err)
	}

	return job, true, nil
}

func (ac *AgentClient) SendHeartbeat() error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/agents/%s/heartbeat", ac.ServerURL, ac.Agent.ID), nil)
	if err != nil {
		return fmt.Errorf("failed to create heartbeat request: %w", err)
	}

	resp, err := ac.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp model.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("heartbeat failed with status: %s, and failed to parse error response: %v", resp.Status, err)
		}

		return fmt.Errorf("heartbeat failed with status: %s, error: %s", resp.Status, errResp.Message)
	}

	return nil
}
