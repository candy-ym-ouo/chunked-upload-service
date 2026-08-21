# Bug Reproduction

## Scenario

The worker loop replaces its caller context with `context.Background()` before invoking work. Cancellation from the worker owner therefore does not reach the work callback.

## Command

`go test ./internal/worker -count=1 -run '^TestMergerRunPropagatesCancellation$'`

## Expected failure

The callback observes a different context instead of the context supplied to `Run`.
