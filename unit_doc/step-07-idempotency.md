# Step 07: Idempotency Logic

## Logic Summary
- Centralize API idempotency decisions into pure functions for lookup, create, and duplicate resolution.
- Explicitly classify store-unavailable cases to support fail-open behavior.

## Design Reasoning
- Pure decision helpers make the API flow easier to read, test, and reuse.
- Keeping store error interpretation in one place avoids inconsistent dedupe behavior.

## Test Command
```sh
go test ./...
```
