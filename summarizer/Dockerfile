FROM golang:1.24 AS builder
LABEL authors="iv"

COPY shared/ /app/shared/

WORKDIR /app/summarizer

COPY summarizer/go.* ./
RUN go mod download

COPY summarizer/ ./


RUN go build -o app ./cmd/main

FROM debian:bookworm-slim
LABEL authors="iv"


COPY --from=builder /app/summarizer/app /app/summarizer/app

ENV CONFIG_PATH=/app/config.yaml
ENV OPENAI_API_KEY=${OPENAI_API_KEY}

CMD ["/app/summarizer/app"]
