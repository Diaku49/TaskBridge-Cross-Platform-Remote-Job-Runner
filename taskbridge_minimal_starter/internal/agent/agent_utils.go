package agent

import (
	"context"
	"log"
	"net/http"
	"strings"
	"taskbridge/internal/executor"
	"taskbridge/internal/model"
	"time"
)

var heartbeatInterval = 10 * time.Second
var validJobTypes = map[model.JobType]struct{}{
	model.JobHTTPCheck:  {},
	model.JobTCPCheck:   {},
	model.JobFileExists: {},
	model.JobChecksum:   {},
	model.JobWriteFile:  {},
	model.JobWait:       {},
}

func HeartbeatLoop(ac *AgentClient, ctx context.Context) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Heartbeat stopped")
			return
		case <-ticker.C:
			if err := ac.SendHeartbeat(); err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
			}
		}
	}
}

func NewRegistry(capabilities []model.JobType) *executor.Registry {
	registry := executor.NewRegistry()
	for _, cap := range capabilities {
		switch cap {
		case model.JobHTTPCheck:
			registry.Register(executor.NewHTTPCheckExecutor(&http.Client{}))
		case model.JobTCPCheck:
			registry.Register(executor.NewTcpCheckExecutor())
		case model.JobFileExists:
			registry.Register(executor.NewFileExistsExecutor())
		case model.JobChecksum:
			registry.Register(executor.NewChecksumExecutor())
		case model.JobWriteFile:
			registry.Register(executor.NewWriteFileExecutor())
		case model.JobWait:
			registry.Register(executor.NewWaitExecutor())
		}
	}
	return registry
}

func parseCapabilities(raw string) []model.JobType {
	parts := strings.Split(raw, ",")

	caps := make([]model.JobType, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		jt := model.JobType(part)
		if _, ok := validJobTypes[jt]; ok {
			caps = append(caps, jt)
		} else {
			log.Print("Warning: Invalid job type in capabilities: ", part)
		}
	}

	return caps
}
