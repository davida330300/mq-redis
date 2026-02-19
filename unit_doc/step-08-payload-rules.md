# Step 08: Payload Rules

## Logic Summary
- Enforce exactly one of inline `payload` or `payload_ref` with required metadata.
- Reject oversized inline payloads and require pointer-based payloads instead.
- Normalize pointer payloads into a consistent JSON envelope for storage/publish.

## Design Reasoning
- Keep payload validation as pure logic to ensure consistent API behavior.
- Require size and hash metadata for pointer payloads to support observability and verification.

## Test Command
```sh
go test ./...
```
