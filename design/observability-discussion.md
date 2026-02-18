# Observability Discussion

## Goals
- Make enqueue, processing, retry, and delivery outcomes visible end-to-end.
- Provide per-tenant visibility without leaking PII.
- Detect backlog growth, retry storms, and partial outages quickly.

## Signals
- Metrics: latency, throughput, error rates, queue lag, retry counts, DLQ counts.
- Logs: structured, correlated by job id and idempotency key.
- Traces: spans across API, Kafka, worker, and retry dispatcher.

## Key Questions To Answer
- Where is time spent between enqueue and delivery?
- Are retries caused by downstream errors or internal failures?
- Which tenants are impacted and how severely?
- Is backlog growth due to Kafka lag or worker throughput limits?

## Tradeoffs
- High-cardinality labels can explode metrics cost; use sampling or aggregation.
- Full payload logging is unsafe; store only hashes or size metadata.
- End-to-end tracing across Kafka requires context propagation in message headers.
