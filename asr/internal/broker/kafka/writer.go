package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kxddry/lectura/asr/internal/config"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
	"time"
)

type Writer struct {
	w *kafka.Writer
}

func CheckAlive(brokers []string) error {
	if len(brokers) == 0 {
		return fmt.Errorf("empty list of brokers")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Controller()
	return err
}

func NewWriter(cfg *config.Kafka) Writer {
	var compression kafka.Compression
	switch cfg.Write.Compression {
	case "gzip":
		compression = kafka.Gzip
	case "snappy":
		compression = kafka.Snappy
	case "lz4":
		compression = kafka.Lz4
	case "zstd":
		compression = kafka.Zstd
	default:
		compression = compress.None
	}

	var requiredAcks kafka.RequiredAcks
	switch cfg.Write.Acks {
	case "0":
		requiredAcks = kafka.RequireNone
	case "1":
		requiredAcks = kafka.RequireOne
	case "all":
		requiredAcks = kafka.RequireAll
	default:
		requiredAcks = kafka.RequireAll
	}

	w := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  cfg.Write.Topic,
		Balancer:               &kafka.RoundRobin{},
		MaxAttempts:            cfg.Write.Retries,
		RequiredAcks:           requiredAcks,
		Async:                  true,
		Compression:            compression,
		WriteTimeout:           cfg.Write.Timeout,
		AllowAutoTopicCreation: false,
	}
	return Writer{w}
}

func (w Writer) Write(ctx context.Context, record transcribed.BrokerRecord) error {
	msgBytes, _ := json.Marshal(record)

	msg := kafka.Message{
		Value: msgBytes,
	}

	return w.w.WriteMessages(ctx, msg)
}
