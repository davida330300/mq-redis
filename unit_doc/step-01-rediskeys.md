# Step 01: Redis Key/TTL Helpers

## Logic Summary
- Centralize Redis key naming (`job:<id>`, `job:data:<id>`, `retry:jobs`, `retry:lock`).
- Define TTL constants for dedupe, job status/data, and DLQ.

## Design Reasoning
- Keep key formats consistent and reusable across packages.
- Encode TTL policy as typed constants to reduce drift and errors.

## Test Command
```sh
go test ./...
```
