# Disaster Recovery And Multi-Region Spec

## Baseline Approach
- Active-active enqueue and processing across regions.
- Each region can accept traffic independently.
- No cross-region synchronous writes on the hot path.

## Objectives
- RPO <= 5 minutes.
- RTO <= 15 minutes.

## Replication
- Kafka topics are mirrored asynchronously across regions.
- Redis data is replicated asynchronously or via periodic snapshots.
- Idempotency keys are valid across regions to prevent duplicate re-enqueue.

## Failover Procedure (Required)
- Health checks detect regional failure and trigger regional traffic drain.
- Global routing shifts traffic away from the failed region.
- Remaining regions continue processing and replay from mirrored Kafka topics.
- Post-failover reconciliation handles duplicates and stale `processing` states.

## Recovery
- When a failed region returns, it rejoins as an active region after backfill.
- Reconcile data before accepting full traffic.
- Controlled reintroduction; no automatic flip-flop.
