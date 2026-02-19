# Step 09: Framework Skeleton

## Logic Summary
- Introduce YAML config loading and validation for API/worker/retry-dispatcher.
- Add Kafka/Redis/Postgres/Saga skeleton packages and interfaces.
- Wire API to Redis + Kafka producer stub and add runnable worker/retry command skeletons.

## Design Reasoning
- Centralized config enables consistent service wiring and future expansion.
- Skeleton packages establish clear boundaries without forcing integration behavior yet.

## Test Command
```sh
go test ./...
```
