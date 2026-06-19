package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"taskbridge/internal/model"
	"time"

	"github.com/google/uuid"
)

func (s *SqliteStore) CreateJob(job model.Job) (model.Job, error) {
	ctx := context.Background()

	job.ID = uuid.NewString()
	job.Status = model.JobPending
	job.CreatedAt = time.Now().UTC()
	job.AttemptCount = 0

	params, err := createJobParams(job)
	if err != nil {
		return model.Job{}, fmt.Errorf("failed converting, err:%v", err)
	}

	row, err := s.q.CreateJob(ctx, params)
	if err != nil {
		return model.Job{}, fmt.Errorf("db failed, err:%v", err)
	}

	return sqliteJobToModel(row)
}
func (s *SqliteStore) ListJobs() ([]model.Job, error) {
	ctx := context.Background()

	rows, err := s.q.ListJobs(ctx)
	if err != nil {
		return []model.Job{}, fmt.Errorf("db failed, err:%v", err)
	}

	return sqliteJobsToModels(rows)
}
func (s *SqliteStore) GetJob(jobID string) (model.Job, bool, error) {
	ctx := context.Background()

	row, err := s.q.GetJob(ctx, jobID)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Job{}, false, nil
		}
		return model.Job{}, false, fmt.Errorf("db failed, err:%v", err)
	}

	job, err := sqliteJobToModel(row)

	return job, true, err
}

func (s *SqliteStore) CancelJob(jobID string) error {
	return nil
}

func (s *SqliteStore) AssignNextJob(agentID string) (model.Job, bool, error) {
	ctx := context.Background()
	now := time.Now().UTC()

	assignParam := assignNextJobParams(agentID, now)

	row, err := s.q.AssignNextJob(ctx, assignParam)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Job{}, false, nil
		}
		return model.Job{}, false, fmt.Errorf("failed assign job, err:%v", err)
	}

	job, err := sqliteJobToModel(row)

	return job, true, err
}

func (s *SqliteStore) CompleteJob(jobID string, status model.JobStatus, logs []string, result map[string]any, errMsg string) error {
	ctx := context.Background()

	row, err := s.q.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed getting job, err:%v", err)
	}

	if row.Status != string(model.JobRunning) {
		return fmt.Errorf("failed completing job, err:%v", "job is not running")
	}

	now := time.Now().UTC()
	nextStatus := status
	var finishedAt *time.Time

	switch status {
	case model.JobSuccess:
		finishedAt = &now
	case model.JobFailed:
		if row.AttemptCount <= row.MaxRetries {
			nextStatus = model.JobRetrying
		} else {
			finishedAt = &now
		}
	default:
		return fmt.Errorf("failed completing job, err:%v", fmt.Sprintf("invalid completion status: %s", status))
	}

	params, err := completeJobUpdateParams(jobID, nextStatus, finishedAt, logs, result, errMsg)
	if err != nil {
		return fmt.Errorf("failed converting complete job params, err:%v", err)
	}

	rowsAffected, err := s.q.CompleteJobUpdate(ctx, params)
	if err != nil {
		return fmt.Errorf("failed updating complete job, err:%v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("failed completing job, err:%v", "job is not running")
	}

	return nil
}
