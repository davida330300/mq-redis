# Disaster Recovery And Multi-Region Discussion

## Goals
- Survive regional failures without losing accepted jobs.
- Provide a predictable failover path and recovery procedure.
- Avoid cross-region synchronous dependencies on the hot path.

## Patterns
- Active-passive: one primary region, standby region for failover.
- Active-active: multiple regions accept writes, require global dedupe.
- Warm standby: async replication of Redis snapshots and Kafka mirror.

## Tradeoffs
- Active-passive is simpler and cheaper but has higher RTO.
- Active-active reduces RTO but requires global idempotency and conflict handling.
- Cross-region synchronous writes reduce availability and increase latency.

## Open Decisions
- Target RPO and RTO.
- Whether to allow multi-region enqueue or keep a single writer.
- How to replicate Redis and Kafka safely and consistently.
