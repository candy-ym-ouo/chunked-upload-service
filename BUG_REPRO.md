# Bug Reproduction

## Scenario

Service error wrapping uses string formatting without preserving the underlying error chain. A repository not-found error cannot be detected with `errors.Is`.

## Command

`go test ./internal/app -count=1 -run '^TestEnsurePreservesNotFound$'`

## Expected failure

The test reports that `errors.Is` does not recognize the repository's not-found sentinel.
