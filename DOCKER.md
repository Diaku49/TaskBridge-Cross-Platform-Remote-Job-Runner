# TaskBridge Docker

Run the assignment-style in-memory stack:

```bash
docker compose --profile memory up --build
```

- API: `http://localhost:8080`
- Dashboard: `http://localhost:5173`

Run the SQLite-backed stack:

```bash
docker compose --profile sqlite up --build
```

- API: `http://localhost:8081`
- Dashboard: `http://localhost:5174`
- Database volume: `taskbridge_taskbridge-sqlite-data`

Reset SQLite data:

```bash
docker compose --profile sqlite down -v
```

The SQLite stack uses the same server binary with `--store sqlite --sqlite-path /data/taskbridge.db`.
