FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/uploader ./cmd/uploader

FROM alpine:3.20
RUN addgroup -S uploader && adduser -S -G uploader uploader
WORKDIR /app
COPY --from=build /out/uploader /app/uploader
COPY --from=build /src/web /app/web
COPY --from=build /src/migrations /app/migrations
RUN mkdir -p /app/data && chown -R uploader:uploader /app
USER uploader
ENV UPLOAD_LISTEN_ADDR=:8080 UPLOAD_STORAGE_ROOT=/app/data
EXPOSE 8080
VOLUME ["/app/data"]
ENTRYPOINT ["/app/uploader"]
