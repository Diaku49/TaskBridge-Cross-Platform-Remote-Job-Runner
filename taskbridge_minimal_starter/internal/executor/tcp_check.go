package executor

import (
	"context"
	"fmt"
	"net"
	"taskbridge/internal/model"
)

type TcpCheckExecutor struct{}

type TCPCheckPayload struct {
	Address string `json:"address"`
}

func NewTcpCheckExecutor() *TcpCheckExecutor {
	return &TcpCheckExecutor{}
}

func (e *TcpCheckExecutor) Type() model.JobType {
	return model.JobTCPCheck
}

func (e *TcpCheckExecutor) Execute(ctx context.Context, job model.Job) Result {
	logs := make([]string, 0, 4)

	var payload TCPCheckPayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "invalid payload",
			Logs:   []string{"failed to decode payload: " + err.Error()},
		}
	}

	logs = append(logs, "decoded payload successfully")

	if err := CheckTCPPayload(&payload); err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "invalid payload",
			Logs:   append(logs, "payload validation failed: "+err.Error()),
		}
	}

	var dialer net.Dialer

	conn, err := dialer.DialContext(ctx, "tcp", payload.Address)
	if err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "TCP connection failed",
			Logs:   append(logs, "failed to connect to TCP address: "+err.Error()),
			Result: map[string]any{
				"address":   payload.Address,
				"reachable": false,
			},
		}
	}
	defer conn.Close()

	logs = append(logs, fmt.Sprintf("successfully connected to TCP address: %s", payload.Address))

	return Result{
		Status: model.JobSuccess,
		Logs:   logs,
		Result: map[string]any{
			"address":   payload.Address,
			"reachable": true,
		},
	}
}

func CheckTCPPayload(p *TCPCheckPayload) error {
	if p.Address == "" {
		return fmt.Errorf("address is required")
	}

	return nil
}
