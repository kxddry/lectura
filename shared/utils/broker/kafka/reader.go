package kafka

import (
	"context"
	"encoding/json"
	kafka2 "github.com/kxddry/lectura/shared/entities/config/kafka"
	"github.com/kxddry/lectura/shared/entities/summarized"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/segmentio/kafka-go"
	"time"
)

type Reader[T uploaded.Record | transcribed.Record | summarized.Record] struct {
	r *kafka.Reader
}

func NewReader[T uploaded.Record | transcribed.Record | summarized.Record](cfg kafka2.ReaderConfig) Reader[T] {
	var startOffset int64
	switch cfg.StartOffset {
	case "earliest":
		startOffset = kafka.FirstOffset
	case "latest":
		startOffset = kafka.LastOffset
	default:
		startOffset = kafka.LastOffset // fallback
	}

	return Reader[T]{
		kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.Brokers,
			GroupID:        cfg.GroupID,
			Topic:          cfg.Topic,
			MinBytes:       cfg.MinBytes,
			MaxBytes:       cfg.MaxBytes,
			CommitInterval: cfg.CommitInterval,
			StartOffset:    startOffset,
		}),
	}
}

func (r Reader[T]) Messages(ctx context.Context) (<-chan T, <-chan error) {
	msgCh := make(chan T)
	errCh := make(chan error)

	go func() {
		defer close(msgCh)
		defer close(errCh)

		for {
			m, err := r.r.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				errCh <- err
				return
			}

			var record T
			if err = json.Unmarshal(m.Value, &record); err != nil {
				continue
			}

			select {
			case <-ctx.Done():
				return
			case msgCh <- record:
				if err = r.r.CommitMessages(ctx, m); err != nil {
					errCh <- err
					return
				}
			}
		}
	}()

	return msgCh, errCh
}

func (r Reader[T]) CheckAlive() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", r.r.Config().Brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Controller()
	return err
}
