FROM golang:1.24 AS builder
LABEL authors="iv"

COPY shared/ /app/shared/

WORKDIR /app/uploader

COPY uploader/go.* ./
RUN go mod download

COPY uploader/ ./


RUN go build -o app ./cmd/main

FROM debian:bookworm-slim
LABEL authors="iv"

RUN apt-get update && apt-get install -y ffmpeg

COPY --from=builder /app/uploader/app /app/uploader/app

ENV CONFIG_PATH=/app/config.yaml

CMD ["/app/uploader/app"]
