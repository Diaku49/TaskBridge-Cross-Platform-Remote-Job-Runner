# TaskBridge

TaskBridge is a small remote job runner written in Go. It has an HTTP server that stores and assigns jobs, and an agent process that registers itself, polls for compatible work, executes jobs, sends heartbeats, and reports results back to the server.

I built it in the same order I would normally grow a service like this: first the server and job API, then agents and assignment, then the in-memory store required by the assignment, then executors and result handling. After that I added a SQLite store behind the same store interface, a simple dashboard, and Docker profiles for both memory and SQLite runs.

## What Is Included

- Go HTTP API server in `taskbridge_minimal_starter/cmd/server`
- Go agent in `taskbridge_minimal_starter/cmd/agent`
- In-memory store used by default
- Optional SQLite store using `modernc.org/sqlite` and sqlc-generated queries
- Executors for:
  - `http_check`
  - `tcp_check`
  - `file_exists`
  - `checksum`
  - `write_file`
  - `wait`
- Static dashboard in `frontend`
- Docker Compose profiles for memory and SQLite modes
- Example JSON payloads and curl commands

## Project Layout

```text
.
├── docker-compose.yml
├── frontend/
│   ├── index.html
│   ├── app.js
│   ├── styles.css
│   └── server.mjs
└── taskbridge_minimal_starter/
    ├── cmd/
    │   ├── agent/
    │   └── server/
    ├── internal/
    │   ├── agent/
    │   ├── api/
    │   ├── executor/
    │   ├── model/
    │   └── store/
    ├── examples/
    └── docs/
```

## Run Locally

Start the server with the default in-memory store:

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

More API examples are in `taskbridge_minimal_starter/docs/API_EXAMPLES.md`.

## SQLite Mode

The server can use SQLite instead of memory:

```bash
cd taskbridge_minimal_starter
go run ./cmd/server \
  --addr :8080 \
  --store sqlite \
  --sqlite-path taskbridge.db
```

The SQLite store initializes its schema on startup. It uses the same API and store interface as memory mode.

## Dashboard

Run the dashboard locally:

```bash
cd frontend
npm run dev
```

Open:

```text
http://127.0.0.1:5173
```

The dashboard shows jobs, agents, status counts, and a job lookup panel. By default it proxies `/api` to `http://localhost:8080`.

## Docker

Run the assignment-style memory stack:

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

Reset the SQLite database:

```bash
docker compose --profile sqlite down -v
```

## Tests

Run the Go checks:

```bash
cd taskbridge_minimal_starter
go test ./...
```

Check Docker Compose configuration:

```bash
docker compose --profile memory config
docker compose --profile sqlite config
```

## Notes

- The memory store is the default path and the core assignment implementation.
- SQLite is optional and selected with `--store sqlite`.
- The dashboard is intentionally small: it is for observing jobs and agents rather than managing every action.
- Job cancellation is still a stretch-goal area; the route is not wired into the server yet.
