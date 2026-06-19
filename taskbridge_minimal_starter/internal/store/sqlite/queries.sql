---- Jobs
-- name: CreateJob :one
INSERT INTO jobs (
    id,
    name,
    job_type,
    payload,
    status,
    created_at,
    max_retries,
    timeout_seconds
) VALUES (
    ?, ?, ?, ?, 'PENDING', ?, ?, ?
)
RETURNING
    id,
    name,
    job_type,
    payload,
    status,
    assigned_agent_id,
    created_at,
    started_at,
    finished_at,
    attempt_count,
    max_retries,
    timeout_seconds,
    logs,
    error,
    result;

-- name: ListJobs :many
SELECT
    id,
    name,
    job_type,
    payload,
    status,
    assigned_agent_id,
    created_at,
    started_at,
    finished_at,
    attempt_count,
    max_retries,
    timeout_seconds,
    logs,
    error,
    result
FROM jobs
ORDER BY created_at ASC, id ASC;

-- name: GetJob :one
SELECT
    id,
    name,
    job_type,
    payload,
    status,
    assigned_agent_id,
    created_at,
    started_at,
    finished_at,
    attempt_count,
    max_retries,
    timeout_seconds,
    logs,
    error,
    result
FROM jobs
WHERE id = ?;

-- name: JobExists :one
SELECT EXISTS (
    SELECT 1
    FROM jobs
    WHERE id = ?
) AS job_exists;

-- name: GetJobCompletionState :one
SELECT
    id,
    status,
    attempt_count,
    max_retries
FROM jobs
WHERE id = ?;

-- name: CompleteJobUpdate :execrows
UPDATE jobs
SET
    status = ?1,

    assigned_agent_id = CASE
        WHEN ?1 = 'RETRYING' THEN NULL
        ELSE assigned_agent_id
    END,

    started_at = CASE
        WHEN ?1 = 'RETRYING' THEN NULL
        ELSE started_at
    END,

    finished_at = CASE
        WHEN ?1 IN ('SUCCESS', 'FAILED') THEN ?2
        ELSE NULL
    END,

    logs = ?3,
    result = ?4,
    error = ?5
WHERE id = ?6
  AND status = 'RUNNING';

-- name: AssignNextJob :one
UPDATE jobs
SET
    status = 'RUNNING',
    assigned_agent_id = ?1,
    attempt_count = attempt_count + 1,
    started_at = ?2
WHERE id = (
    SELECT j.id
    FROM jobs j
    JOIN agent_capabilities ac
        ON ac.job_type = j.job_type
    WHERE ac.agent_id = ?1
      AND j.status IN ('PENDING', 'RETRYING')
    ORDER BY j.created_at ASC, j.id ASC
    LIMIT 1
)
RETURNING
    id,
    name,
    job_type,
    payload,
    status,
    assigned_agent_id,
    created_at,
    started_at,
    finished_at,
    attempt_count,
    max_retries,
    timeout_seconds,
    logs,
    error,
    result;


---- Agents

-- name: CreateAgent :one
INSERT INTO agents (
    id,
    hostname,
    os,
    arch,
    version,
    last_seen
) VALUES (
    ?, ?, ?, ?, ?, ?
)
RETURNING
    id,
    hostname,
    os,
    arch,
    version,
    last_seen;

-- name: AddAgentCapability :exec
INSERT INTO agent_capabilities (
    agent_id,
    job_type
) VALUES (
    ?, ?
);

-- name: AgentExists :one
SELECT EXISTS (
    SELECT 1
    FROM agents
    WHERE id = ?
) AS agent_exists;

-- name: ListAgents :many
SELECT
    a.id,
    a.hostname,
    a.os,
    a.arch,
    a.version,
    a.last_seen,
    COALESCE(GROUP_CONCAT(ac.job_type), '') AS capabilities
FROM agents a
LEFT JOIN agent_capabilities ac
    ON ac.agent_id = a.id
GROUP BY
    a.id,
    a.hostname,
    a.os,
    a.arch,
    a.version,
    a.last_seen
ORDER BY a.last_seen DESC, a.id ASC;

-- name: UpdateAgentHeartbeat :execrows
UPDATE agents
SET last_seen = ?
WHERE id = ?;
