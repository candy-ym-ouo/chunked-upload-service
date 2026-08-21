# Bug Reproduction

## Scenario

Completion starts reading chunks at index one and requests an out-of-range final index. The repository also reports a fabricated placeholder chunk, corrupting completeness checks.

## Command

`go test ./internal/app -count=1 -run '^TestCompleteAllUploadedChunks$'`

## Expected failure

Completion of a session whose only real chunk is index zero fails instead of creating the file.
