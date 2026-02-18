# Observability Spec

## Metrics (Required)
- API enqueue latency and status rate (success, error).
- Worker processing latency and status rate (done, retrying, dlq).
- End-to-end delivery latency (enqueue to done).
- Kafka consumer lag and throughput.
- Retry queue depth and retry rate.
- DLQ rate and DLQ backlog.
- Redis error rate and latency.

## Logs (Required Fields)
- `job_id`, `idempotency_key`, `tenant_id`, `attempt`, `status`.
- `kafka.topic`, `kafka.partition`, `kafka.offset`.
- `region`, `instance_id`, `component`.
- No raw payloads or secrets; log payload size and hash only.

## Tracing
- Propagate trace context in Kafka message headers.
- Each job should have a root span from API enqueue to worker completion.
- Retry attempts must link to the original trace via span links.

## Alerts (Minimum Set)
- Enqueue error rate exceeds threshold.
- End-to-end delivery latency breaches SLO.
- Retry queue depth grows continuously.
- DLQ rate spikes.
- Kafka consumer lag grows beyond expected bounds.
