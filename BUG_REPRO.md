# Bug Reproduction

## Scenario

Download registration uses `context.Background()` instead of the request context. A canceled request is therefore not observable by the repository layer.

## Command

`go test ./internal/app -count=1 -run '^TestRegisterDownloadUsesRequestContext$'`

## Expected failure

The repository callback does not receive the canceled context and the test reports that cancellation was ignored.
