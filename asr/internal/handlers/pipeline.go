package handlers

import (
	"context"
	"fmt"
	"github.com/kxddry/lectura/asr/internal/config"
	"github.com/kxddry/lectura/asr/internal/whisper"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/kxddry/lectura/shared/utils/broker/kafka"
	"io"
)

type s3client interface {
	Download(ctx context.Context, bucket string, key string) (io.ReadCloser, error)
}

func Pipeline(ctx context.Context, cfg config.Config, cli s3client, kp kafka.Pipeline[uploaded.Record, transcribed.Record], msg uploaded.Record) error {
	file, err := cli.Download(ctx, msg.Bucket, msg.UUID+".wav")
	if err != nil {
		return err
	}

	resp, err := whisper.CallWhisperAPI(cfg.WhisperAPI, file)
	if err != nil || resp.Text == "" {
		return fmt.Errorf("callWhisperAPI: %w", err)
	}

	if err = kp.W.Write(ctx, transcribed.Record{UUID: msg.UUID, Text: resp.Text, Language: resp.Language}); err != nil {
		return fmt.Errorf("upload text: %w", err)
	}
	return nil
}
