# Step 04: Minimal API Core (Gin)

## Logic Summary
- Accept `POST /jobs` with `idempotency_key` and `payload`.
- Validate input, enforce payload size, and return consistent responses.
- Dedupe via store lookup; generate job IDs on first enqueue.
- Fail-open when store is unavailable and publish with a warning.

## Design Reasoning
- Keep the API handler small and pure, with interfaces for store and producer.
- Enforce input boundaries early to avoid downstream errors.
- Align fail-open behavior with the availability requirement.

## Test Command
```sh
go test ./...
```
