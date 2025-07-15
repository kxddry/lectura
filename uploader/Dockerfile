FROM golang:1.24 AS builder
LABEL authors="iv"
LABEL org.opencontainers.image.source=https://github.com/kxddry/lectura

COPY shared/ /app/shared/

WORKDIR /app/uploader

COPY uploader/go.* ./
RUN go mod download

COPY uploader/ ./


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/main

FROM debian:bookworm-slim
LABEL authors="iv"

RUN apt-get update && apt-get install -y ffmpeg

COPY --from=builder /app/uploader/app /app/uploader/app


CMD ["/app/uploader/app"]
