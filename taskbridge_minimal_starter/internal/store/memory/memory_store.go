package memory

import (
	"fmt"
	"sync"
	"taskbridge/internal/model"
	"time"

	"github.com/google/uuid"
)

// TODO: Candidate should implement MemoryStore with mutex-safe maps.
type MemoryStore struct {
	mu                sync.RWMutex
	agents            map[string]model.Agent
	jobs              map[string]model.Job
	jobsIDs           []string
	assignableJobsIDs map[model.JobType][]string
}

func NewMemoryStore() *MemoryStore {
	ms := MemoryStore{
		mu:                sync.RWMutex{},
		agents:            make(map[string]model.Agent),
		jobs:              make(map[string]model.Job),
		jobsIDs:           make([]string, 0),
		assignableJobsIDs: make(map[model.JobType][]string),
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
	job.CreatedAt = time.Now().UTC()
	job.AttemptCount = 0

	ms.jobsIDs = append(ms.jobsIDs, id)
	ms.assignableJobsIDs[job.Type] = append(ms.assignableJobsIDs[job.Type], job.ID)
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
	if !ok {
		return model.Job{}, false, nil
	}

	return job, ok, nil
}

func (ms *MemoryStore) CancelJob(jobID string) error {
	return nil
}

func (ms *MemoryStore) CompleteJob(jobID string, status model.JobStatus, logs []string, result map[string]any, errMsg string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	job, ok := ms.jobs[jobID]
	if !ok {
		return fmt.Errorf("job not found")
	}

	if job.Status != model.JobRunning {
		return fmt.Errorf("job is not running")
	}

	now := time.Now().UTC()

	switch status {
	case model.JobSuccess:
		job.Status = model.JobSuccess
		job.FinishedAt = &now

	case model.JobFailed:
		if job.AttemptCount <= job.MaxRetries {
			job.Status = model.JobRetrying
			job.AssignedAgentID = ""
			job.StartedAt = nil

			ms.assignableJobsIDs[job.Type] = append(ms.assignableJobsIDs[job.Type], job.ID)
		} else {
			job.Status = model.JobFailed
			job.FinishedAt = &now
		}

	default:
		return fmt.Errorf("invalid completion status: %s", status)
	}

	job.Logs = logs
	job.Result = result
	job.Error = errMsg
	ms.jobs[jobID] = job

	return nil
}

// Agents
func (ms *MemoryStore) RegisterAgent(agent model.Agent) (model.Agent, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, ok := ms.agents[agent.ID]; ok {
		return model.Agent{}, fmt.Errorf("Agent with this ID already exist")
	}

	agent.LastSeen = time.Now().UTC()
	agent.Status = model.OnlineAgent

	ms.agents[agent.ID] = agent

	return agent, nil
}

func (ms *MemoryStore) ListAgents() ([]model.Agent, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	now := time.Now().UTC()
	agents := make([]model.Agent, 0, len(ms.agents))
	for _, agent := range ms.agents {
		if now.Sub(agent.LastSeen) > model.AgentOfflineAfter {
			agent.Status = model.OfflineAgent
		} else {
			agent.Status = model.OnlineAgent
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

func (ms *MemoryStore) Heartbeat(agentID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	agent, ok := ms.agents[agentID]
	if !ok {
		return fmt.Errorf("agent not found")
	}
	agent.LastSeen = time.Now().UTC()
	agent.Status = model.OnlineAgent
	ms.agents[agentID] = agent

	return nil
}

func (ms *MemoryStore) AssignNextJob(agentID string) (model.Job, bool, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	agent, ok := ms.agents[agentID]
	if !ok {
		return model.Job{}, false, fmt.Errorf("agent not found")
	}

	for _, cap := range agent.Capabilities {
		ids := ms.assignableJobsIDs[cap]

		for i := 0; i < len(ids); i++ {
			jobID := ids[i]

			job, ok := ms.jobs[jobID]
			if !ok {
				// remove stale queue entry
				ids = append(ids[:i], ids[i+1:]...)
				i--
				continue
			}

			if job.Status != model.JobPending && job.Status != model.JobRetrying {
				// remove stale queue entry
				ids = append(ids[:i], ids[i+1:]...)
				i--
				continue
			}

			now := time.Now().UTC()

			job.Status = model.JobRunning
			job.AssignedAgentID = agentID
			job.AttemptCount++
			job.StartedAt = &now

			ms.jobs[job.ID] = job

			// remove assigned job from pending queue
			ids = append(ids[:i], ids[i+1:]...)
			ms.assignableJobsIDs[cap] = ids

			return job, true, nil
		}

		ms.assignableJobsIDs[cap] = ids
	}

	return model.Job{}, false, nil
}
