FROM golang:1.24 AS builder
LABEL authors="iv"
LABEL org.opencontainers.image.source=https://github.com/kxddry/lectura

COPY shared/ /app/shared/

WORKDIR /app/updater

COPY updater/go.* ./
RUN go mod download

COPY updater/ ./


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/main

FROM scratch
LABEL authors="iv"

COPY --from=builder /app/updater/app /app/updater/app


CMD ["/app/updater/app"]
