FROM golang:1.24 AS builder
LABEL authors="iv"

COPY shared/ /app/shared/

WORKDIR /app/updater

COPY updater/go.* ./
RUN go mod download

COPY updater/ ./


RUN go build -o app ./cmd/main

FROM scratch
LABEL authors="iv"

COPY --from=builder /app/updater/app /app/updater/app

ENV CONFIG_PATH=/app/config.yaml

CMD ["/app/updater/app"]
