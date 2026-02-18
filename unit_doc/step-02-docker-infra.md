# Step 02: Docker Infra (Postgres/Redis/Kafka)

## Logic Summary
- Provide local infrastructure services for persistence and messaging.

## Design Reasoning
- Use Docker Compose to standardize local environments.
- Keep services minimal and aligned with project components.

## Test Command
```sh
# No unit tests (infra config only)
# Optional smoke check:
# docker compose up -d
```
