# Step 13: Retry Scheduler Logic

## Logic Summary
- Implement exponential backoff with jitter and max cap.
- Provide pure helpers to compute retry delay and Redis ZSET score.

## Design Reasoning
- Pure functions keep retry math deterministic and easy to test.
- Jitter reduces thundering herds while honoring max delay.

## Test Command
```sh
go test ./...
```
