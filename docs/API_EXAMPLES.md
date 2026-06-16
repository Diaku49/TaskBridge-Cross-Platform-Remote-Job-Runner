# API Examples

These examples assume the server is running locally:

```bash
go run ./cmd/server --addr :8080
```

Use this shell variable to keep commands shorter:

```bash
SERVER=http://localhost:8080
```

## Health

```bash
curl -sS "$SERVER/health"
```

## Jobs

Create an HTTP check job:

```bash
curl -sS -X POST "$SERVER/jobs" \
  -H "Content-Type: application/json" \
  -d @examples/create-http-check-job.json
```

Create a wait job:

```bash
curl -sS -X POST "$SERVER/jobs" \
  -H "Content-Type: application/json" \
  -d @examples/create-wait-job.json
```

Create a TCP check job:

```bash
curl -sS -X POST "$SERVER/jobs" \
  -H "Content-Type: application/json" \
  -d @examples/create-tcp-check-job.json
```

Create a file-exists job:

```bash
mkdir -p /tmp/taskbridge-demo
printf "taskbridge demo file\n" > /tmp/taskbridge-demo/input.txt

curl -sS -X POST "$SERVER/jobs" \
  -H "Content-Type: application/json" \
  -d @examples/create-file-exists-job.json
```

Create a checksum job:

```bash
mkdir -p /tmp/taskbridge-demo
printf "taskbridge demo file\n" > /tmp/taskbridge-demo/input.txt

curl -sS -X POST "$SERVER/jobs" \
  -H "Content-Type: application/json" \
  -d @examples/create-checksum-job.json
```

Create a write-file job:

```bash
curl -sS -X POST "$SERVER/jobs" \
  -H "Content-Type: application/json" \
  -d @examples/create-write-file-job.json
```

List jobs:

```bash
curl -sS "$SERVER/jobs"
```

Get one job:

```bash
JOB_ID=<job-id>
curl -sS "$SERVER/jobs/$JOB_ID"
```

Submit a job result manually. This endpoint expects the job to already be assigned and `RUNNING`; in the normal flow, the agent submits this request for you.

```bash
JOB_ID=<job-id>
curl -sS -X POST "$SERVER/jobs/$JOB_ID/result" \
  -H "Content-Type: application/json" \
  -d @examples/submit-job-result-success.json
```

## Agents

Register an agent manually:

```bash
curl -sS -X POST "$SERVER/agents/register" \
  -H "Content-Type: application/json" \
  -d @examples/register-agent.json
```

Send a heartbeat:

```bash
AGENT_ID=agent-dev-1
curl -sS -X POST "$SERVER/agents/$AGENT_ID/heartbeat"
```

Poll for the next compatible job:

```bash
AGENT_ID=agent-dev-1
curl -i -X POST "$SERVER/agents/$AGENT_ID/next-job"
```

List agents:

```bash
curl -sS "$SERVER/agents"
```

## Agent Demo

In one terminal, run the server:

```bash
go run ./cmd/server --addr :8080
```

In another terminal, run an agent with all implemented capabilities:

```bash
go run ./cmd/agent \
  --server http://localhost:8080 \
  --id agent-dev-1 \
  --capabilities http_check,tcp_check,file_exists,checksum,write_file,wait
```

Then create a job from another terminal:

```bash
curl -sS -X POST "$SERVER/jobs" \
  -H "Content-Type: application/json" \
  -d @examples/create-http-check-job.json
```

Check the job list after the agent polls and reports the result:

```bash
curl -sS "$SERVER/jobs"
```
