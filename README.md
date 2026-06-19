# TaskBridge

TaskBridge is a small remote job runner written in Go. It has an HTTP API server that stores and assigns jobs, and an agent process that registers itself, polls for compatible work, executes jobs, sends heartbeats, and reports results back to the server.

The project was built in stages: server and job API first, then agents and assignment, then the required in-memory store, executors, result handling, optional SQLite persistence, a small dashboard, and Docker profiles for both storage modes.

## What Is Included

- Go HTTP API server in `taskbridge_minimal_starter/cmd/server`
- Go agent in `taskbridge_minimal_starter/cmd/agent`
- In-memory store used by default
- Optional SQLite store using `modernc.org/sqlite` and sqlc-generated queries
- Executors for `http_check`, `tcp_check`, `file_exists`, `checksum`, `write_file`, and `wait`
- Static dashboard in `frontend`
- Docker Compose profiles for memory and SQLite modes
- API examples in `taskbridge_minimal_starter/docs/API_EXAMPLES.md`

## Project Layout

```text
.
├── docker-compose.yml
├── DOCKER.md
├── frontend/
│   ├── index.html
│   ├── app.js
│   ├── styles.css
│   └── server.mjs
└── taskbridge_minimal_starter/
    ├── cmd/
    │   ├── agent/
    │   └── server/
    ├── docs/
    ├── examples/
    └── internal/
        ├── agent/
        ├── api/
        ├── executor/
        ├── model/
        └── store/
```

## Run Locally

Start the API server with the default in-memory store:

```bash
cd taskbridge_minimal_starter
go run ./cmd/server --addr :8080
```

Start an agent in another terminal:

```bash
cd taskbridge_minimal_starter
go run ./cmd/agent \
  --server http://localhost:8080 \
  --id agent-dev-1 \
  --capabilities http_check,tcp_check,file_exists,checksum,write_file,wait
```

Create a job:

```bash
cd taskbridge_minimal_starter
curl -sS -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d @examples/create-wait-job.json
```

List jobs:

```bash
curl -sS http://localhost:8080/jobs
```

## SQLite Mode

SQLite uses the same API and store interface as memory mode:

```bash
cd taskbridge_minimal_starter
go run ./cmd/server \
  --addr :8080 \
  --store sqlite \
  --sqlite-path taskbridge.db
```

The schema is initialized automatically when the SQLite store starts.

## Dashboard

Run the dashboard locally:

```bash
cd frontend
npm run dev
```

Open `http://127.0.0.1:5173`.

The dashboard shows jobs, agents, status counts, job details, and a create-job form. It calls the backend API directly, so the Go server's CORS middleware allows browser requests from the dashboard port.

Set `TASKBRIDGE_API` if the API is running somewhere else:

```bash
TASKBRIDGE_API=http://localhost:8081 npm run dev
```

## Docker

Run the in-memory stack:

```bash
docker compose --profile memory up --build
```

- API: `http://localhost:8080`
- Dashboard: `http://localhost:5173`

Run the SQLite stack:

```bash
docker compose --profile sqlite up --build
```

- API: `http://localhost:8081`
- Dashboard: `http://localhost:5174`
- SQLite data volume: `taskbridge_taskbridge-sqlite-data`

When creating `http_check` jobs from the Docker dashboard, the generated payload targets the API service name such as `http://api-memory:8080/health`. That address is reachable from the agent container. From local, non-Docker runs, `http://localhost:8080/health` is fine.

Reset the SQLite database:

```bash
docker compose --profile sqlite down -v
```

More Docker notes are in `DOCKER.md`.

## Checks

Run the Go package checks:

```bash
cd taskbridge_minimal_starter
go test ./...
```

Check Docker Compose configuration:

```bash
docker compose --profile memory config
docker compose --profile sqlite config
```

## Current Scope

- Memory mode is the main assignment path.
- SQLite mode is optional and selected with `--store sqlite`.
- The dashboard is intentionally small: it is for observing agents/jobs and creating simple jobs.
- Job cancellation is not wired into the server routes.
