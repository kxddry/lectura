package handlers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kxddry/lectura/shared/entities/summarized"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/kxddry/lectura/summarizer/internal/config"
	"github.com/kxddry/lectura/summarizer/internal/entities"
	"io"
)

type Client interface {
	Uploader
	Downloader
}

type Uploader interface {
	Upload(ctx context.Context, fc uploaded.FileConfig) error
}

type Downloader interface {
	Download(ctx context.Context, fc uploaded.FileConfig) (io.ReadCloser, error)
}

type Writer interface {
	Write(ctx context.Context, record summarized.BrokerRecord) error
}

type KafkaPipeline struct {
	InputCh <-chan transcribed.BrokerRecord
	ErrCh   <-chan error

	W Writer
}

type MessageSender interface {
	SendMessage(msg []byte) (entities.ChatResponse, error)
}

func Pipeline(ctx context.Context, cfg *config.Config, sender MessageSender, c Client, kp KafkaPipeline, msg transcribed.BrokerRecord) error {
	const op = "handlers.Pipeline"

	fc := uploaded.FileConfig{
		Extension: ".txt",
		FileName:  msg.TextID,
		FileID:    msg.TextID,
		File:      nil,
		FileSize:  msg.TextSize,
		Bucket:    cfg.Storage.BucketInput,
		FileType:  "text/plain",
	}

	file, err := c.Download(ctx, fc)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer file.Close()

	txtBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	resp, err := sender.SendMessage(txtBytes)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	txt := resp.Choices[0].Message.Content
	if len(txt) == 0 {
		return fmt.Errorf("%s: empty response", op)
	}

	rec := summarized.BrokerRecord{
		TextName: msg.TextName,
		TextID:   msg.TextID,
		TextSize: int64(len(txt)),
		Language: msg.Language,
	}

	err = c.Upload(ctx, uploaded.FileConfig{
		Extension: ".txt",
		FileName:  msg.TextName,
		FileID:    msg.TextID,
		File:      io.NopCloser(bytes.NewReader(txt)),
		FileSize:  int64(len(txt)),
		Bucket:    cfg.Storage.BucketOutput,
		FileType:  "text/plain",
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = kp.W.Write(ctx, rec)
	if err != nil {
		return fmt.Errorf("%s failed to write in kafka: %w", op, err)
	}

	return nil
}
