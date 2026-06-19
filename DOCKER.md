# TaskBridge Docker

The Compose file has two profiles: `memory` and `sqlite`. Both start the Go API, one agent, and the dashboard.

## Memory Stack

```bash
docker compose --profile memory up --build
```

- API: `http://localhost:8080`
- Dashboard: `http://localhost:5173`
- Store: in-memory

## SQLite Stack

```bash
docker compose --profile sqlite up --build
```

- API: `http://localhost:8081`
- Dashboard: `http://localhost:5174`
- Store: SQLite at `/data/taskbridge.db`
- Database volume: `taskbridge_taskbridge-sqlite-data`

Reset SQLite data:

```bash
docker compose --profile sqlite down -v
```

## API URLs And Job URLs

The dashboard calls the backend directly from the browser:

- Memory dashboard API URL: `http://localhost:8080`
- SQLite dashboard API URL: `http://localhost:8081`

Generated job payload examples use Docker service names:

- Memory job target: `http://api-memory:8080`
- SQLite job target: `http://api-sqlite:8080`

That split is intentional. The browser runs on your host machine, but the agent executes jobs inside Docker.
