FROM golang:1.24 AS builder
LABEL authors="iv"
LABEL org.opencontainers.image.source=https://github.com/kxddry/lectura

COPY shared/ /app/shared/

WORKDIR /app/summarizer

COPY summarizer/go.* ./
RUN go mod download

COPY summarizer/ ./


RUN go build -o app ./cmd/main

FROM scratch
LABEL authors="iv"


COPY --from=builder /app/summarizer/app /app/summarizer/app


CMD ["/app/summarizer/app"]
