FROM golang:1.24 AS builder
LABEL authors="iv"

COPY shared/ /app/shared/

WORKDIR /app/updater

COPY updater/go.* ./
RUN go mod download

COPY updater/ ./


RUN go build -o app ./cmd/main

FROM debian:bookworm-slim
LABEL authors="iv"

RUN apt-get update && apt-get install -y ffmpeg

COPY --from=builder /app/updater/app /app/updater/app

ENV CONFIG_PATH=/app/config.yaml

CMD ["/app/updater/app"]
