# Repository Guidelines

## Project Overview

TaskBridge is a Go 1.22 starter project for a cross-platform remote job runner. It has two binaries:

- `cmd/server`: starts the HTTP API server.
- `cmd/agent`: registers an agent and is intended to poll, execute, heartbeat, and submit results.

Core packages live under `internal/`:

- `internal/model`: shared domain models and JSON DTOs.
- `internal/store`: `Store` interface and mutex-backed `MemoryStore`.
- `internal/api`: HTTP handlers and JSON response helpers.
- `internal/agent`: agent client, heartbeat loop, capability parsing, executor registry setup.
- `internal/executor`: executor interface, registry, payload decoding, and job-type executors.
- `internal/config`: placeholder config structs.

The repository is part starter skeleton and part in-progress implementation. Prefer completing the existing interfaces and patterns before adding new architecture.

## Commands

- Run all compile checks/tests: `go test ./...`
- Run the server: `go run ./cmd/server --addr :8080`
- Run the agent: `go run ./cmd/agent --server http://localhost:8080 --id agent-dev-1`
- Format Go code before finishing edits: `gofmt -w <changed .go files>`

There are currently no `_test.go` files, so `go test ./...` is primarily a compile/package check until tests are added.

## Implementation Notes

- Keep store operations concurrency-safe with the existing mutex pattern in `MemoryStore`.
- Preserve job lifecycle fields consistently: `CreatedAt`, `StartedAt`, `FinishedAt`, `AttemptCount`, `AssignedAgentID`, `Status`, `Logs`, `Error`, and `Result`.
- The server uses Go 1.22 `net/http` method-aware mux patterns such as `POST /jobs`.
- JSON responses should go through `HTTPResponse` and `HTTPError` in `internal/api/server.go`.
- Executors should decode payloads with `DecodePayload`, validate using per-executor helpers, honor `context.Context`, return `executor.Result`, and avoid unsafe filesystem or network behavior beyond the requested job.
- Agent capabilities are parsed in `internal/agent/agent_utils.go`; supported job types are defined in `internal/model/model.go`.

## Current Caveats

- `internal/executor/tcp_check.go` has a local uncommitted change adding `CheckTCPPayload`; do not discard it.
- API and agent methods are not fully aligned yet. For example, the agent sends `POST` for heartbeat and next-job while the server registers `PUT` heartbeat and `GET` next-job.
- `SubmitJobResult` currently reads `r.PathValue("jobsId")`, but the route uses `{jobId}`.
- `CancelJob` is a stub and the cancel route is not currently registered.
- `ChecksumExecutor` is still a stub. Other executors vary in completeness.
- README and inline TODOs still describe some starter-state work that is now partially implemented.

## Style

- Use idiomatic, small Go functions and keep package boundaries simple.
- Prefer standard library solutions unless the project already depends on a package.
- Add focused tests when changing lifecycle, API, store, agent, or executor behavior.
- Avoid broad refactors while assignment features are still being completed.
