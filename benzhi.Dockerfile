FROM golang:1.22

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build ./...

ENV UPLOAD_LISTEN_ADDR=:8080
ENV UPLOAD_STORAGE_ROOT=/app/data
EXPOSE 8080
CMD ["go", "run", "./cmd/uploader"]
