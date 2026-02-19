# Step 11: E2E Smoke Test

## Logic Summary
- Add an E2E smoke test that starts Docker Compose, launches the API, and verifies Redis/Kafka/Postgres integration.
- Validate that API enqueue writes Redis keys and publishes to Kafka.

## Design Reasoning
- Provides a minimal integration check while core worker logic is still pending.
- Uses build tags to keep the test opt-in.

## Test Command
```sh
go test -tags e2e ./e2e
```
