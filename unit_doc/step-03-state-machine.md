# Step 03: Job State Machine

## Logic Summary
- Define job states and their allowed transitions.
- Provide helper functions to validate transitions and detect terminal states.

## Design Reasoning
- Centralize state rules so worker/API logic can rely on a single source of truth.
- Keep transitions explicit to prevent accidental illegal moves.

## Test Command
```sh
go test ./...
```
