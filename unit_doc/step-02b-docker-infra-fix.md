# Step 02b: Docker Infra Fix (Apache Kafka)

## Logic Summary
- Replace Kafka image with Apache Kafka and update KRaft env vars.

## Design Reasoning
- The previous Bitnami tag was not resolvable in Docker Hub.
- Apache Kafka image is official and provides consistent KRaft configuration.

## Test Command
```sh
# No unit tests (infra config only)
# Optional smoke check:
# docker compose up -d
```
