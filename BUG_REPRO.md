# Bug Reproduction

## Scenario

Chunk sorting reuses the caller's slice and shortens its range before sorting. The result can lose the final chunk and mutate shared slice storage.

## Command

`go test ./internal/domain -count=1 -run '^TestSortChunksPreservesInput$'`

## Expected failure

The focused test detects that sorting does not preserve the complete input slice.
