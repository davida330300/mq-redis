# Consistency And Degradation Rules

## Sequencing Rules
- API must write idempotency status before publishing to Kafka.
- API order: `SETNX job:<id>=queued` -> `SET job:data:<id>` -> publish Kafka `jobs`.
- If Kafka publish fails, API returns error and the client retries with the same idempotency key.
- Worker uses Kafka auto-commit; Redis updates are best-effort and must be idempotent.
- Worker order on success: set `processing` -> execute handler -> set `done` (offset may commit independently).
- Worker order on failure: set `retrying` -> schedule retry (offset may commit independently).
- DLQ order: publish to `jobs.dlq` and set status `dlq`; offsets may commit independently.

## Degradation Policy
- Redis unavailable at API: fail-open and publish to Kafka; return 202 with a warning flag to indicate dedupe may be degraded.
- Redis unavailable at Worker: log and retry Redis write when possible; auto-commit may still advance offsets, so reconciliation is required.
- Kafka unavailable: API returns 503; workers back off and retry consume or publish.

## Locking And Claiming
- Retry dispatcher uses a short-lived Redis lock with TTL to reduce duplicate requeue.
- Due jobs are claimed atomically from `retry:jobs` (Lua or equivalent) before re-publish.
- Duplicate requeue is acceptable; idempotency at worker handles duplicates.

## Reconciliation
- A sweeper periodically requeues stale `queued` jobs (older than publish timeout).
- Stuck `processing` jobs beyond a max processing time are moved to `retrying`.
- If `job:data:<id>` is missing, mark job `dlq` or alert for manual repair.
