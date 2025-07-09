FROM golang:1.24 AS builder
LABEL authors="iv"

WORKDIR /app

COPY asr/ ./asr
COPY shared/ ./shared

WORKDIR /app/asr

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/main

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/asr/app .

ENV CONFIG_PATH=/app/config.yaml

CMD ["./app"]
