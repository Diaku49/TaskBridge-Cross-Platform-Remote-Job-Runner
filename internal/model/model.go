package model

import "time"

// JobStatus represents the lifecycle state of a job.
type JobStatus string

const (
	JobPending  JobStatus = "PENDING"
	JobRunning  JobStatus = "RUNNING"
	JobRetrying JobStatus = "RETRYING"
	JobSuccess  JobStatus = "SUCCESS"
	JobFailed   JobStatus = "FAILED"
	JobCanceled JobStatus = "CANCELED"
)

// JobType represents supported job execution types.
type JobType string

const (
	JobHTTPCheck  JobType = "http_check"
	JobTCPCheck   JobType = "tcp_check"
	JobFileExists JobType = "file_exists"
	JobChecksum   JobType = "checksum"
	JobWriteFile  JobType = "write_file"
	JobWait       JobType = "wait"
)

// Job is the main server-side job entity.
type Job struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Type            JobType        `json:"type"`
	Payload         map[string]any `json:"payload"`
	Status          JobStatus      `json:"status"`
	AssignedAgentID string         `json:"assigned_agent_id,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	StartedAt       *time.Time     `json:"started_at,omitempty"`
	FinishedAt      *time.Time     `json:"finished_at,omitempty"`
	AttemptCount    int            `json:"attempt_count"`
	MaxRetries      int            `json:"max_retries"`
	TimeoutSeconds  int            `json:"timeout_seconds"`
	Logs            []string       `json:"logs,omitempty"`
	Error           string         `json:"error,omitempty"`
	Result          map[string]any `json:"result,omitempty"`
}

// Agent is the server-side representation of a connected worker.
type Agent struct {
	ID           string    `json:"id"`
	Hostname     string    `json:"hostname"`
	OS           string    `json:"os"`
	Arch         string    `json:"arch"`
	Version      string    `json:"version"`
	Capabilities []JobType `json:"capabilities"`
	LastSeen     time.Time `json:"last_seen"`
	Status       string    `json:"status"`
}

var (
	OnlineAgent       = "online"
	OfflineAgent      = "offline"
	AgentOfflineAfter = 30 * time.Second
)

// TODO: Candidate should define request/response DTOs clearly.

// Job DTOs
type CreateJobRequest struct {
	Name           string         `json:"name" validate:"required,notblank"`
	Type           JobType        `json:"type" validate:"required,oneof=http_check tcp_check file_exists checksum write_file wait"`
	Payload        map[string]any `json:"payload" validate:"required"`
	MaxRetries     int            `json:"max_retries" validate:"gte=0"`
	TimeoutSeconds int            `json:"timeout_seconds" validate:"gte=0"`
}

type SubmitJobResultRequest struct {
	Status JobStatus      `json:"status" validate:"required,oneof=SUCCESS FAILED"`
	Logs   []string       `json:"logs"`
	Result map[string]any `json:"result,omitempty"`
	Error  string         `json:"error,omitempty"`
}

// Agent DTOs
type RegisterAgentRequest struct {
	ID           string    `json:"id" validate:"required,notblank"`
	Hostname     string    `json:"hostname" validate:"required,notblank"`
	OS           string    `json:"os" validate:"required,notblank"`
	Arch         string    `json:"arch" validate:"required,notblank"`
	Version      string    `json:"version" validate:"required,notblank"`
	Capabilities []JobType `json:"capabilities" validate:"required,min=1,dive,oneof=http_check tcp_check file_exists checksum write_file wait"`
}

// Not sure
type ErrorResponse struct {
	Message string `json:"message"`
}
type NextJobResponse struct {
	Job Job `json:"job"`
}

// Suggested DTOs:
//   CreateJobRequest
//   RegisterAgentRequest
//   JobResultRequest
//   ErrorResponse
//   NextJobResponse
