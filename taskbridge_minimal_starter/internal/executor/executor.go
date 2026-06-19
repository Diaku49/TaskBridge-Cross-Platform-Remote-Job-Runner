package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"taskbridge/internal/model"
)

// Result is returned after executing a job.
type Result struct {
	Status model.JobStatus `json:"status"`
	Logs   []string        `json:"logs"`
	Result map[string]any  `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// Executor executes a single job type.
type Executor interface {
	Type() model.JobType
	Execute(ctx context.Context, job model.Job) Result
}

// Registry maps job types to executors.
type Registry struct {
	executors map[model.JobType]Executor
}

func NewRegistry() *Registry {
	return &Registry{executors: map[model.JobType]Executor{}}
}

func (r *Registry) Register(ex Executor) {
	r.executors[ex.Type()] = ex
}

func (r *Registry) Get(t model.JobType) (Executor, bool) {
	ex, ok := r.executors[t]
	return ex, ok
}

func DecodePayload(payload map[string]any, dst any) error {
	jsonByte, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal failed, err:%v", err)
	}

	return json.Unmarshal(jsonByte, dst)
}
