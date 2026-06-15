package executor

import (
	"context"
	"taskbridge/internal/model"
)

type WaitExecutor struct{}

func NewWaitExecutor() *WaitExecutor {
	return &WaitExecutor{}
}

func (e *WaitExecutor) Type() model.JobType {
	return model.JobWait
}

func (e *WaitExecutor) Execute(ctx context.Context, job model.Job) Result {

	return Result{}
}
