# Bug Reproduction

## Scenario

An empty `UPLOAD_STORAGE_ROOT` disables the runnable default storage path, and the size policy rejects a file exactly equal to `MaxFileSize`.

## Command

`go test ./internal/config -count=1 -run '^TestLoadHasRunnableDefaults$'`

## Expected failure

The default configuration fails validation because the storage root is empty.
