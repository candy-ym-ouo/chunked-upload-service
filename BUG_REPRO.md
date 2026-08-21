# Bug Reproduction

## Scenario

Concurrent file download counter updates use a read lock while writing the in-memory file map. Run the focused regression test with the race detector.

## Command

`go test -race ./internal/store/postgres -count=1 -run '^TestSaveFileConcurrentUpdates$'`

## Expected failure

The test reports an unsafe concurrent map write/read or a data race in `SaveFile`.
