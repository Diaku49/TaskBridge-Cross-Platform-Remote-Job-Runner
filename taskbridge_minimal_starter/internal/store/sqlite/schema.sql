PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS job_types (
    name TEXT PRIMARY KEY NOT NULL
);

CREATE TABLE IF NOT EXISTS agents (
    id TEXT PRIMARY KEY NOT NULL,
    hostname TEXT NOT NULL,
    os TEXT NOT NULL,
    arch TEXT NOT NULL,
    version TEXT NOT NULL,
    last_seen TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_capabilities (
    agent_id TEXT NOT NULL,
    job_type TEXT NOT NULL,

    PRIMARY KEY (agent_id, job_type),

    FOREIGN KEY (agent_id)
        REFERENCES agents(id)
        ON DELETE CASCADE,

    FOREIGN KEY (job_type)
        REFERENCES job_types(name)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY NOT NULL,

    name TEXT NOT NULL,

    job_type TEXT NOT NULL,
    payload TEXT NOT NULL,

    status TEXT NOT NULL CHECK (
        status IN (
            'PENDING',
            'RUNNING',
            'RETRYING',
            'SUCCESS',
            'FAILED',
            'CANCELED'
        )
    ),

    assigned_agent_id TEXT,

    created_at TEXT NOT NULL,
    started_at TEXT,
    finished_at TEXT,

    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    max_retries INTEGER NOT NULL DEFAULT 0 CHECK (max_retries >= 0),
    timeout_seconds INTEGER NOT NULL CHECK (timeout_seconds >= 0),

    logs TEXT NOT NULL DEFAULT '[]',
    error TEXT NOT NULL DEFAULT '',
    result TEXT,

    FOREIGN KEY (job_type)
        REFERENCES job_types(name),

    FOREIGN KEY (assigned_agent_id)
        REFERENCES agents(id)
        ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_jobs_assignable_type_created
ON jobs(job_type, created_at)
WHERE status IN ('PENDING', 'RETRYING');

INSERT OR IGNORE INTO job_types (name) VALUES
    ('http_check'),
    ('tcp_check'),
    ('file_exists'),
    ('checksum'),
    ('write_file'),
    ('wait');
