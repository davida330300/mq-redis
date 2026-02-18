# Consistency And Degradation Rules

## Sequencing Rules
- API must write idempotency status before publishing to Kafka.
- API order: `SETNX job:<id>=queued` -> `SET job:data:<id>` -> publish Kafka `jobs`.
- If Kafka publish fails, API returns error and the client retries with the same idempotency key.
- Worker must only commit Kafka offsets after Redis state is updated.
- Worker order on success: set `processing` -> execute handler -> set `done` -> commit offset.
- Worker order on failure: set `retrying` -> schedule retry -> commit offset.
- DLQ order: publish to `jobs.dlq` and set status `dlq`; commit only after both succeed.

## Degradation Policy
- Redis unavailable at API: fail-open and publish to Kafka; return 202 with a warning flag to indicate dedupe may be degraded.
- Redis unavailable at Worker: pause processing or retry Redis write; do not commit until Redis is updated.
- Kafka unavailable: API returns 503; workers back off and retry consume or publish.

## Locking And Claiming
- Retry dispatcher uses a short-lived Redis lock with TTL to reduce duplicate requeue.
- Due jobs are claimed atomically from `retry:jobs` (Lua or equivalent) before re-publish.
- Duplicate requeue is acceptable; idempotency at worker handles duplicates.

## Reconciliation
- A sweeper periodically requeues stale `queued` jobs (older than publish timeout).
- Stuck `processing` jobs beyond a max processing time are moved to `retrying`.
- If `job:data:<id>` is missing, mark job `dlq` or alert for manual repair.
