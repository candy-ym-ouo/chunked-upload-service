# Bug Reproduction

## Scenario

After validating a chunk's exact size, the application passes one byte less to disk storage. An exact-size upload is rejected as a content length mismatch.

## Command

`go test ./internal/app -count=1 -run '^TestPutChunkAcceptsExactSize$'`

## Expected failure

The exact-size chunk upload returns `content length mismatch` instead of succeeding.
