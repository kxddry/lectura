FROM golang:1.24 AS builder
LABEL authors="iv"

COPY shared/ /app/shared/

WORKDIR /app/asr

COPY asr/go.* ./
RUN go mod download

COPY asr/ ./


RUN go build -o app ./cmd/main

FROM scratch
LABEL authors="iv"


COPY --from=builder /app/asr/app /app/asr/app

ENV CONFIG_PATH=/app/config.yaml

CMD ["/app/asr/app"]
