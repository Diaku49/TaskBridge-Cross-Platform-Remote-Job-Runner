# TaskBridge Dashboard

Static dashboard for the TaskBridge API.

```bash
npm run dev
```

By default the dashboard runs at `http://localhost:5173` and calls the API at `http://localhost:8080`.

To target another API:

```bash
TASKBRIDGE_API=http://localhost:8081 npm run dev
```

`TASKBRIDGE_API` is written into `config.js` when the dashboard starts, so the browser calls that API URL directly.

`TASKBRIDGE_JOB_TARGET` can be set separately for generated job payload examples. In Docker Compose this uses service names like `http://api-memory:8080`, because the agent executes the job from inside the Docker network.
