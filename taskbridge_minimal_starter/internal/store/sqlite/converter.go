package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"taskbridge/internal/model"
	"taskbridge/internal/store/sqlite/generated"
	"time"
)

const sqliteTimeFormat = time.RFC3339Nano

func sqliteJobToModel(row generated.Job) (model.Job, error) {
	var payload map[string]any
	if err := parseJSONText(row.Payload, &payload); err != nil {
		return model.Job{}, fmt.Errorf("decode job payload: %w", err)
	}

	var logs []string
	if err := parseJSONText(row.Logs, &logs); err != nil {
		return model.Job{}, fmt.Errorf("decode job logs: %w", err)
	}

	var result map[string]any
	if err := parseNullableJSONText(row.Result, &result); err != nil {
		return model.Job{}, fmt.Errorf("decode job result: %w", err)
	}

	createdAt, err := parseTime(row.CreatedAt)
	if err != nil {
		return model.Job{}, fmt.Errorf("decode job created_at: %w", err)
	}

	startedAt, err := parseNullableTime(row.StartedAt)
	if err != nil {
		return model.Job{}, fmt.Errorf("decode job started_at: %w", err)
	}

	finishedAt, err := parseNullableTime(row.FinishedAt)
	if err != nil {
		return model.Job{}, fmt.Errorf("decode job finished_at: %w", err)
	}

	return model.Job{
		ID:              row.ID,
		Name:            row.Name,
		Type:            model.JobType(row.JobType),
		Payload:         payload,
		Status:          model.JobStatus(row.Status),
		AssignedAgentID: row.AssignedAgentID.String,
		CreatedAt:       createdAt,
		StartedAt:       startedAt,
		FinishedAt:      finishedAt,
		AttemptCount:    int(row.AttemptCount),
		MaxRetries:      int(row.MaxRetries),
		TimeoutSeconds:  int(row.TimeoutSeconds),
		Logs:            logs,
		Error:           row.Error,
		Result:          result,
	}, nil
}

func sqliteJobsToModels(rows []generated.Job) ([]model.Job, error) {
	jobs := make([]model.Job, 0, len(rows))
	for _, row := range rows {
		job, err := sqliteJobToModel(row)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func createJobParams(job model.Job) (generated.CreateJobParams, error) {
	payload := job.Payload
	if payload == nil {
		payload = map[string]any{}
	}

	payloadText, err := jsonText(payload)
	if err != nil {
		return generated.CreateJobParams{}, fmt.Errorf("encode job payload: %w", err)
	}

	return generated.CreateJobParams{
		ID:             job.ID,
		Name:           job.Name,
		JobType:        string(job.Type),
		Payload:        payloadText,
		CreatedAt:      formatTime(job.CreatedAt),
		MaxRetries:     int64(job.MaxRetries),
		TimeoutSeconds: int64(job.TimeoutSeconds),
	}, nil
}

func sqliteAgentToModel(row generated.Agent, capabilities []model.JobType) (model.Agent, error) {
	lastSeen, err := parseTime(row.LastSeen)
	if err != nil {
		return model.Agent{}, fmt.Errorf("decode agent last_seen: %w", err)
	}

	return model.Agent{
		ID:           row.ID,
		Hostname:     row.Hostname,
		OS:           row.Os,
		Arch:         row.Arch,
		Version:      row.Version,
		Capabilities: capabilities,
		LastSeen:     lastSeen,
		Status:       agentStatus(lastSeen),
	}, nil
}

func listAgentRowToModel(row generated.ListAgentsRow) (model.Agent, error) {
	capabilities, err := parseCapabilities(row.Capabilities)
	if err != nil {
		return model.Agent{}, err
	}

	return sqliteAgentToModel(generated.Agent{
		ID:       row.ID,
		Hostname: row.Hostname,
		Os:       row.Os,
		Arch:     row.Arch,
		Version:  row.Version,
		LastSeen: row.LastSeen,
	}, capabilities)
}

func listAgentRowsToModels(rows []generated.ListAgentsRow) ([]model.Agent, error) {
	agents := make([]model.Agent, 0, len(rows))
	for _, row := range rows {
		agent, err := listAgentRowToModel(row)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}
	return agents, nil
}

func registerAgentParams(agent model.Agent) generated.CreateAgentParams {
	return generated.CreateAgentParams{
		ID:       agent.ID,
		Hostname: agent.Hostname,
		Os:       agent.OS,
		Arch:     agent.Arch,
		Version:  agent.Version,
		LastSeen: formatTime(agent.LastSeen),
	}
}

func addAgentCapabilityParams(agentID string, jobType model.JobType) generated.AddAgentCapabilityParams {
	return generated.AddAgentCapabilityParams{
		AgentID: agentID,
		JobType: string(jobType),
	}
}

func updateAgentHeartbeatParams(agentID string, lastSeen time.Time) generated.UpdateAgentHeartbeatParams {
	return generated.UpdateAgentHeartbeatParams{
		ID:       agentID,
		LastSeen: formatTime(lastSeen),
	}
}

func assignNextJobParams(agentID string, startedAt time.Time) generated.AssignNextJobParams {
	return generated.AssignNextJobParams{
		AssignedAgentID: nullableString(agentID),
		StartedAt:       nullableTime(&startedAt),
	}
}

func completeJobUpdateParams(jobID string, status model.JobStatus, finishedAt *time.Time, logs []string, result map[string]any, errMsg string) (generated.CompleteJobUpdateParams, error) {
	logsText, err := jsonText(logs)
	if err != nil {
		return generated.CompleteJobUpdateParams{}, fmt.Errorf("encode job logs: %w", err)
	}

	resultText, err := nullableJSONText(result)
	if err != nil {
		return generated.CompleteJobUpdateParams{}, fmt.Errorf("encode job result: %w", err)
	}

	return generated.CompleteJobUpdateParams{
		ID:         jobID,
		Status:     string(status),
		FinishedAt: nullableTime(finishedAt),
		Logs:       logsText,
		Result:     resultText,
		Error:      errMsg,
	}, nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format(sqliteTimeFormat)
}

func parseTime(value string) (time.Time, error) {
	return time.Parse(sqliteTimeFormat, value)
}

func nullableTime(t *time.Time) sql.NullString {
	if t == nil || t.IsZero() {
		return sql.NullString{}
	}
	return sql.NullString{
		String: formatTime(*t),
		Valid:  true,
	}
}

func parseNullableTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid || value.String == "" {
		return nil, nil
	}

	t, err := parseTime(value.String)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func nullableString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{
		String: value,
		Valid:  true,
	}
}

func jsonText(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func nullableJSONText(value map[string]any) (sql.NullString, error) {
	if value == nil {
		return sql.NullString{}, nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return sql.NullString{}, err
	}

	return sql.NullString{
		String: string(data),
		Valid:  true,
	}, nil
}

func parseJSONText(value string, dst any) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return json.Unmarshal([]byte(value), dst)
}

func parseNullableJSONText(value sql.NullString, dst any) error {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil
	}
	return json.Unmarshal([]byte(value.String), dst)
}

func parseCapabilities(value any) ([]model.JobType, error) {
	if value == nil {
		return nil, nil
	}

	raw, ok := value.(string)
	if !ok {
		if bytes, ok := value.([]byte); ok {
			raw = string(bytes)
		} else {
			return nil, fmt.Errorf("unexpected capabilities type %T", value)
		}
	}

	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	capabilities := make([]model.JobType, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		capabilities = append(capabilities, model.JobType(part))
	}

	return capabilities, nil
}

func agentStatus(lastSeen time.Time) string {
	if time.Since(lastSeen) > model.AgentOfflineAfter {
		return model.OfflineAgent
	}
	return model.OnlineAgent
}
