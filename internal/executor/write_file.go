package executor

import (
	"context"
	"taskbridge/internal/model"
)

type WriteFileExecutor struct{}

func NewWriteFileExecutor() *WriteFileExecutor {
	return &WriteFileExecutor{}
}

func (e *WriteFileExecutor) Type() model.JobType {
	return model.JobWriteFile
}

func (e *WriteFileExecutor) Execute(ctx context.Context, job model.Job) Result {

	return Result{}
}
