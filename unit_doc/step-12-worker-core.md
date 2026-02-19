# Step 12: Worker Core (Auto-Commit)

## Logic Summary
- Consume Kafka messages with auto-commit and update Redis job status.
- Track attempts in Redis, schedule a single retry, then send to DLQ.

## Design Reasoning
- Auto-commit keeps the worker simple; idempotency + reconciliation handle drift.
- Redis attempt counters allow deterministic retry/DLQ decisions.

## Test Command
```sh
go test ./...
```
