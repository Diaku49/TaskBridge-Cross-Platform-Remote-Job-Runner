package sqlite

import (
	"context"
	"fmt"
	"taskbridge/internal/model"
	"time"
)

func (s *SqliteStore) RegisterAgent(agent model.Agent) (model.Agent, error) {
	ctx := context.Background()

	agent.LastSeen = time.Now().UTC()
	agent.Status = model.OnlineAgent

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Agent{}, fmt.Errorf("failed begin register agent transaction, err:%v", err)
	}
	defer tx.Rollback()

	q := s.q.WithTx(tx)
	params := registerAgentParams(agent)
	row, err := q.CreateAgent(ctx, params)
	if err != nil {
		return model.Agent{}, fmt.Errorf("failed creating agent, err:%v", err)
	}

	for _, capability := range agent.Capabilities {
		if err := q.AddAgentCapability(ctx, addAgentCapabilityParams(agent.ID, capability)); err != nil {
			return model.Agent{}, fmt.Errorf("failed adding agent capability, err:%v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return model.Agent{}, fmt.Errorf("failed commit register agent transaction, err:%v", err)
	}

	createdAgent, err := sqliteAgentToModel(row, agent.Capabilities)
	if err != nil {
		return model.Agent{}, fmt.Errorf("failed converting agent, err:%v", err)
	}

	return createdAgent, nil
}

func (s *SqliteStore) Heartbeat(agentID string) error {
	ctx := context.Background()
	now := time.Now().UTC()

	rowsAffected, err := s.q.UpdateAgentHeartbeat(ctx, updateAgentHeartbeatParams(agentID, now))
	if err != nil {
		return fmt.Errorf("failed updating agent heartbeat, err:%v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("failed updating agent heartbeat, err:%v", "agent not found")
	}

	return nil
}

func (s *SqliteStore) ListAgents() ([]model.Agent, error) {
	ctx := context.Background()

	rows, err := s.q.ListAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed listing agents, err:%v", err)
	}

	agents, err := listAgentRowsToModels(rows)
	if err != nil {
		return nil, fmt.Errorf("failed converting agents, err:%v", err)
	}

	return agents, nil
}
