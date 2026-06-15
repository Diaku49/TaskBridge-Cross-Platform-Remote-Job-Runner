package executor

import (
	"context"
	"taskbridge/internal/model"
)

type TcpCheckExecutor struct{}

func NewTcpCheckExecutor() *TcpCheckExecutor {
	return &TcpCheckExecutor{}
}

func (e *TcpCheckExecutor) Type() model.JobType {
	return model.JobTCPCheck
}

func (e *TcpCheckExecutor) Execute(ctx context.Context, job model.Job) Result {

	return Result{}
}
