# Chunked Upload Service

Run with `go run ./cmd/uploader`. The default server listens on `:8080` and stores data under `./data`. Configuration uses `UPLOAD_LISTEN_ADDR`, `UPLOAD_STORAGE_ROOT`, `UPLOAD_REAPER_INTERVAL`, and `UPLOAD_BEARER_TOKEN`.

The service exposes session creation, resumable chunk uploads, checksum-verified completion, file listing/content download, cancellation, and health checks. PostgreSQL migration files are in `migrations/`; the default executable uses a thread-safe repository implementation so local validation needs no external service.
