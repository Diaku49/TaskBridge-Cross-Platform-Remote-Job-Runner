# TaskBridge Dashboard

Small static dashboard for the TaskBridge API. It can list jobs and agents, show status counts, inspect one job, and create simple jobs.

## Run Locally

```bash
npm run dev
```

Open `http://localhost:5173`.

By default the dashboard calls the API at `http://localhost:8080`.

## Configuration

Use `TASKBRIDGE_API` when the browser should call a different API URL:

```bash
TASKBRIDGE_API=http://localhost:8081 npm run dev
```

Use `TASKBRIDGE_JOB_TARGET` when generated job payload examples should point somewhere different from the browser API URL:

```bash
TASKBRIDGE_JOB_TARGET=http://api-memory:8080 npm run dev
```

This matters in Docker because the browser calls published host ports like `http://localhost:8080`, while the agent executes jobs inside the Docker network and can reach services by names like `http://api-memory:8080`.

The frontend server writes these values into `/config.js` when it starts. It does not proxy API requests.
