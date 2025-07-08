FROM golang:1.24 AS builder
LABEL authors="iv"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/main

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app .
ENV CONFIG_PATH=/app/config.yaml

CMD ["./app"]