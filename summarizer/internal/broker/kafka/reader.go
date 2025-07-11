package kafka

import (
	"context"
	"encoding/json"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/summarizer/internal/config"
	"github.com/segmentio/kafka-go"
	"time"
)

type Reader struct {
	r *kafka.Reader
}

func NewReader(cfg *config.Kafka) Reader {
	var startOffset int64
	switch cfg.Read.StartOffset {
	case "earliest":
		startOffset = kafka.FirstOffset
	case "latest":
		startOffset = kafka.LastOffset
	default:
		startOffset = kafka.LastOffset // fallback
	}

	return Reader{
		kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.Brokers,
			GroupID:        cfg.Read.GroupID,
			Topic:          cfg.Read.Topic,
			MinBytes:       cfg.Read.MinBytes,
			MaxBytes:       cfg.Read.MaxBytes,
			CommitInterval: cfg.Read.CommitInterval,
			StartOffset:    startOffset,
		}),
	}
}

func (r *Reader) Messages(ctx context.Context) (<-chan transcribed.BrokerRecord, <-chan error) {
	msgCh := make(chan transcribed.BrokerRecord)
	errCh := make(chan error, 1)

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

			var record transcribed.BrokerRecord
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

func (r *Reader) CheckAlive() error {
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
