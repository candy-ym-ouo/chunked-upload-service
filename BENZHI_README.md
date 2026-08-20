# chunked-upload-service

可靠分片上传服务，支持断点续传、分片及整体 SHA-256 校验、流式合并、文件下载、过期回收、健康检查和指标接口。

## 标准命令

```bash
go build ./...
go test ./...
go test -race ./...
go vet ./...
go run ./cmd/uploader
```

## Docker

```bash
./build_benzhi_docker.sh chunked-upload-service linux/amd64
./build_benzhi_docker.sh chunked-upload-service linux/arm64
docker run --rm -it -p 8080:8080 -v upload-data:/app/data chunked-upload-service
```

健康检查：`GET /api/v1/healthz`；就绪检查：`GET /api/v1/readyz`。默认监听 `:8080`，数据目录为 `./data`。
