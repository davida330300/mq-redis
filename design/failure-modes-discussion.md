# Failure Modes And Multi-Node Behavior

## Worker Crash Or Restart
- Failure: consumer process dies after fetch; offsets may auto-commit.
- Expected: duplicates are tolerated via Redis status checks; reconciliation should detect stuck or missing states.
- Notes: processing must be idempotent; status writes are authoritative for dedupe.

## Redis Unavailable
- Failure: API cannot write `job:*` keys.
- Expected: API fails closed (503) and does not publish to Kafka to preserve idempotency.
- Notes: clients retry with same idempotency key; optional sweeper can requeue `queued` jobs.

## Kafka Unavailable
- Failure: API cannot publish; worker cannot consume.
- Expected: API returns 503; job remains `queued` in Redis and is safe to retry.
- Notes: backlog grows; recovery is by producer retry or a requeue sweeper.

## Retry Dispatcher Contention
- Failure: multiple nodes pick the same due job.
- Expected: claim uses a short-lived lock and atomic ZSET pop to avoid duplicates.
- Notes: duplicates are tolerated by idempotency checks.

## Network Partition
- Failure: components have partial connectivity (Kafka ok, Redis down or vice versa).
- Expected: workers log and retry Redis writes when possible; offsets may advance due to auto-commit.
- Notes: reconciliation job should repair drift (stale `processing`, missing `queued` republishes).

## DLQ Under Partition
- Failure: DLQ publish succeeds but Redis update fails, or vice versa.
- Expected: retry DLQ publish until both Kafka and Redis reflect `dlq`.
- Notes: duplicates in DLQ are acceptable; downstream should dedupe by job id.
