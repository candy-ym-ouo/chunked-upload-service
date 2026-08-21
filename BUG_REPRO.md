# Bug Reproduction

## Scenario

Remaining byte calculations subtract total size from uploaded bytes instead of subtracting uploaded bytes from total size. This makes resume status report an incorrect remaining amount.

## Command

`go test ./internal/app -count=1 -run '^TestBuildResumeReportsRemainingBytes$'`

## Expected failure

For a ten-byte upload with five bytes present, the reported remaining amount is not five bytes.
