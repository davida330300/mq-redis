# Step 05: Runnable API Server (cmd/api)

## Logic Summary
- Provide a runnable Gin server on `:8080` with `POST /jobs`.
- Add `GET /healthz` for basic health checks.
- Use in-memory Store/Producer implementations for local runs.

## Design Reasoning
- Keep dependencies minimal while enabling a runnable API.
- Separate in-memory implementations under `internal/*/memory` for easy replacement.

## Test Command
```sh
go test ./...
```
