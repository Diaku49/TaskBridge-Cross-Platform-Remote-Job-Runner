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
			// case model.JobTCPCheck:
			// 	registry.Register(&executor.TCPCheckExecutor{})
			// case model.JobFileExists:
			// 	registry.Register(&executor.FileExistsExecutor{})
			// case model.JobChecksum:
			// 	registry.Register(&executor.ChecksumExecutor{})
			// case model.JobWriteFile:
			// 	registry.Register(&executor.WriteFileExecutor{})
			// case model.JobWait:
			// 	registry.Register(&executor.WaitExecutor{})
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
