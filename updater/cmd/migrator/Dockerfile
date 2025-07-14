FROM golang:1.24 AS builder
LABEL authors="iv"

COPY shared/ /app/shared/
WORKDIR /app/migrator


COPY updater/go.* ./
RUN go mod download

COPY updater/ ./

RUN go build -o app ./cmd/migrator

FROM debian:bookworm-slim
LABEL authors="iv"

COPY --from=builder /app/migrator/app /app/app
COPY migrations/ /app/migrations

ENV OPERATION=${OPERATION}
ENV CONFIG_PATH=/app/config.yaml

CMD ["/app/app"]
