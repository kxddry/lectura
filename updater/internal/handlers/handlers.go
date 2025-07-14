package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/kxddry/lectura/shared/entities/summarized"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/shared/entities/uploaded"
)

type Storage interface {
	AddFile(ctx context.Context, msg uploaded.Record) error
	AddTranscription(ctx context.Context, msg transcribed.Record) error
	AddSummarization(ctx context.Context, msg summarized.Record) error
	UpdateFile(ctx context.Context, uuid string, status int) error
}

const (
	upload = iota
	transcribe
	summarize
)

func ProcessMessage(ctx context.Context, msg any, s Storage) error {
	const op = "handlers.ProcessMessage"
	var err error
	switch msg.(type) {
	case uploaded.Record:
		err = s.AddFile(ctx, msg.(uploaded.Record))
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case transcribed.Record:
		msgg := msg.(transcribed.Record)
		err = s.AddTranscription(ctx, msgg)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		_ = s.UpdateFile(ctx, msgg.UUID, transcribe)
	case summarized.Record:
		msgg := msg.(summarized.Record)
		err = s.AddSummarization(ctx, msgg)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		_ = s.UpdateFile(ctx, msgg.UUID, summarize)

	default:
		panic(fmt.Errorf("%s: %w", op, errors.New("unsupported message type")))
	}
	return nil
}
