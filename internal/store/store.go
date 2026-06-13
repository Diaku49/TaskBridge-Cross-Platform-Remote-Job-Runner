package store

import (
	"sync"
	"taskbridge/internal/model"
	"time"

	"github.com/google/uuid"
)

// Store defines the required persistence operations.
// Candidate should first implement an in-memory store, then optionally add SQLite.
type Store interface {
	CreateJob(job model.Job) (model.Job, error)
	ListJobs() ([]model.Job, error)
	GetJob(jobID string) (model.Job, bool, error)
	CancelJob(jobID string) error

	RegisterAgent(agent model.Agent) (model.Agent, error)
	Heartbeat(agentID string) error
	ListAgents() ([]model.Agent, error)

	AssignNextJob(agentID string, capabilities []model.JobType) (model.Job, bool, error)
	CompleteJob(jobID string, status model.JobStatus, logs []string, result map[string]any, errMsg string) error
}

// TODO: Candidate should implement MemoryStore with mutex-safe maps.
type MemoryStore struct {
	mu             sync.RWMutex
	agents         map[string]model.Agent
	jobs           map[string]model.Job
	jobsIDs        []string
	pendingJobsIDs []string
}

func NewMemoryStore() *MemoryStore {
	ms := MemoryStore{
		mu:             sync.RWMutex{},
		agents:         make(map[string]model.Agent),
		jobs:           make(map[string]model.Job),
		jobsIDs:        make([]string, 0),
		pendingJobsIDs: make([]string, 0, 20),
	}

	return &ms
}

// Jobs
func (ms *MemoryStore) CreateJob(job model.Job) (model.Job, error) {
	id := uuid.NewString()

	ms.mu.Lock()
	defer ms.mu.Unlock()

	job.ID = id
	job.Status = model.JobPending
	job.CreatedAt = time.Now()
	job.AttemptCount = 0

	ms.jobsIDs = append(ms.jobsIDs, id)
	ms.pendingJobsIDs = append(ms.pendingJobsIDs, id)
	ms.jobs[job.ID] = job

	return job, nil
}

func (ms *MemoryStore) ListJobs() ([]model.Job, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	jobs := make([]model.Job, 0, len(ms.jobs))
	for _, id := range ms.jobsIDs {
		jobs = append(jobs, ms.jobs[id])
	}

	return jobs, nil
}

func (ms *MemoryStore) GetJob(jobId string) (model.Job, bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	job, ok := ms.jobs[jobId]

	return job, ok, nil
}

func (ms *MemoryStore) CancelJob(jobID string) error {
	return nil
}

// func (ms *MemoryStore) CancelJob(jobID string) error

func (ms *MemoryStore) CompleteJob(jobID string, status model.JobStatus, logs []string, result map[string]any, errMsg string) error {

	return nil
}

// Agents
func (ms *MemoryStore) RegisterAgent(agent model.Agent) (model.Agent, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	agent.LastSeen = time.Now()
	agent.Status = "active"

	ms.agents[agent.ID] = agent

	return agent, nil
}

func (ms *MemoryStore) ListAgents() ([]model.Agent, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	agents := make([]model.Agent, 0, len(ms.agents))
	for _, agent := range ms.agents {
		agents = append(agents, agent)
	}

	return agents, nil
}

func (ms *MemoryStore) Heartbeat(agentID string) error {
	return nil
}

func (ms *MemoryStore) AssignNextJob(agentID string, capabilities []model.JobType) (model.Job, bool, error) {
	return model.Job{}, false, nil
}

// TODO: Stretch goal: SQLite-backed Store implementation.
