# Step 06: Redis Store Implementation

## Logic Summary
- Implement the API Store using Redis with atomic create via WATCH/MULTI.
- Store idempotency key, job status, and job payload with TTLs.
- Map Redis errors to `ErrStoreUnavailable` for fail-open behavior.

## Design Reasoning
- WATCH/MULTI keeps `idem:<key>` + job keys consistent without Lua.
- Explicit TTLs keep Redis bounded and align with SLO expectations.
- Errors are normalized to keep API behavior consistent.

## Test Command
```sh
go test ./...
```
