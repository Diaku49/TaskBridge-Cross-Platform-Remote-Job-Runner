package executor

import (
	"context"
	"taskbridge/internal/model"
)

type ChecksumExecutor struct{}

func NewChecksumExecutor() *ChecksumExecutor {
	return &ChecksumExecutor{}
}

func (e *ChecksumExecutor) Type() model.JobType {
	return model.JobChecksum
}

func (e *ChecksumExecutor) Execute(ctx context.Context, job model.Job) Result {

	return Result{}
}
