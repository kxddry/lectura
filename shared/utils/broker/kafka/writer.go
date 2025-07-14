package kafka

import (
	"context"
	"encoding/json"
	kafka2 "github.com/kxddry/lectura/shared/entities/config/kafka"
	"github.com/kxddry/lectura/shared/entities/summarized"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
	"time"
)

type Writer[T uploaded.Record | transcribed.Record | summarized.Record] struct {
	w *kafka.Writer
}

func (w Writer[T]) Write(ctx context.Context, record T) error {
	msgBytes, err := json.Marshal(record)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Value: msgBytes,
	}
	return w.w.WriteMessages(ctx, msg)
}

func NewWriter[T uploaded.Record | transcribed.Record | summarized.Record](cfg kafka2.WriterConfig) Writer[T] {
	var compression kafka.Compression

	switch cfg.Compression {
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
	switch cfg.Acks {
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
		Topic:                  cfg.Topic,
		Balancer:               &kafka.RoundRobin{},
		MaxAttempts:            cfg.Retries,
		RequiredAcks:           requiredAcks,
		Async:                  true,
		Compression:            compression,
		WriteTimeout:           cfg.Timeout,
		AllowAutoTopicCreation: false,
	}
	return Writer[T]{w: w}
}

func (w Writer[T]) CheckAlive(brokers []string) error {
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
