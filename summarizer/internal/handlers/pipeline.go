package handlers

import (
	"context"
	"fmt"
	"github.com/kxddry/lectura/shared/entities/summarized"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	kafka2 "github.com/kxddry/lectura/shared/utils/broker/kafka"
	"github.com/kxddry/lectura/summarizer/internal/entities"
)

type MessageSender interface {
	SendMessage(msg []byte, lang string) (entities.ChatResponse, error)
}

func Pipeline[R transcribed.Record, W summarized.Record](
	ctx context.Context, sender MessageSender, kp kafka2.Pipeline[R, W], msg transcribed.Record) error {
	const op = "handlers.Pipeline"

	txt := msg.Text
	lang := msg.Language

	txtBytes := []byte(txt)

	resp, err := sender.SendMessage(txtBytes, lang)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	outTxt := resp.Choices[0].Message.Content
	if len(outTxt) == 0 {
		return fmt.Errorf("%s: empty response", op)
	}

	record := summarized.Record{
		UUID: msg.UUID,
		Text: string(outTxt),
	}

	err = kp.W.Write(ctx, W(record))
	if err != nil {
		return fmt.Errorf("%s failed to write in kafka: %w", op, err)
	}

	return nil
}
