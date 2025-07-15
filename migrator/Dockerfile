FROM golang:1.24 AS builder
LABEL authors="iv"

COPY shared/ /shared/
WORKDIR /app


COPY migrator/go.* ./
RUN go mod download

COPY migrator/ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

FROM scratch 
LABEL authors="iv"

COPY --from=builder /app/app /app/app

CMD ["/app/app"]
