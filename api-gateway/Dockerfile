FROM golang:1.24 AS builder
LABEL authors="iv"
LABEL org.opencontainers.image.source=https://github.com/kxddry/lectura

COPY shared/ /app/shared/

WORKDIR /app/api-gateway

COPY api-gateway/go.* ./
RUN go mod download

COPY api-gateway/ ./


RUN go build -o app ./cmd/main

FROM scratch
LABEL authors="iv"


COPY --from=builder /app/api-gateway/app /app/api-gateway/app


CMD ["/app/api-gateway/app"]
