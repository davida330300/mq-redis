# Step 10: Real Connect Checks

## Logic Summary
- Add Kafka/Redis/Postgres connectivity checks with timeouts and warn-on-failure behavior.
- Wire API to a real Kafka producer implementation via kafka-go.

## Design Reasoning
- Early connectivity checks surface misconfigurations without blocking service startup.
- kafka-go keeps the Kafka integration pure Go and easy to embed.

## Test Command
```sh
go test ./...
```
