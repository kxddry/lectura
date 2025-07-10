FROM golang:1.24 AS builder
LABEL authors="iv"

WORKDIR /app

COPY uploader/ ./uploader
COPY shared/ ./shared

WORKDIR /app/uploader

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/main

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/uploader/app .

ENV CONFIG_PATH=/app/config.yaml

CMD ["./app"]
