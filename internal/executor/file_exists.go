package executor

import (
	"context"
	"taskbridge/internal/model"
)

type FileExistsExecutor struct{}

func NewFileExistsExecutor() *FileExistsExecutor {
	return &FileExistsExecutor{}
}

func (e *FileExistsExecutor) Type() model.JobType {
	return model.JobFileExists
}

func (e *FileExistsExecutor) Execute(ctx context.Context, job model.Job) Result {

	return Result{}
}
