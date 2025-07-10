package handlers

import (
	"context"
	"fmt"
	"github.com/kxddry/lectura/asr/internal/config"
	"github.com/kxddry/lectura/asr/internal/whisper"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"io"
	"strings"
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
	Write(ctx context.Context, record transcribed.BrokerRecord) error
}

type KafkaPipeline struct {
	InputCh <-chan uploaded.BrokerRecord
	ErrCh   <-chan error

	W Writer
}

func Pipeline(ctx context.Context, cfg *config.Config, c Client, kp KafkaPipeline, msg uploaded.BrokerRecord) error {
	file, err := c.Download(ctx, uploaded.FileConfig{
		Extension: msg.Extension,
		FileName:  msg.FileName,
		FileID:    msg.FileID,
		File:      nil,
		FileSize:  0,
		Bucket:    cfg.Storage.BucketInput,
		FileType:  msg.FileType,
	})

	resp, err := whisper.CallWhisperAPI(cfg.WhisperAPI, file)
	if err != nil {
		return fmt.Errorf("callWhisperAPI: %w", err)
	}

	err = c.Upload(ctx, uploaded.FileConfig{
		Extension: ".txt",
		FileName:  msg.FileName,
		FileID:    msg.FileID,
		File:      io.NopCloser(strings.NewReader(resp.Text)),
		FileSize:  int64(len(resp.Text)),
		Bucket:    cfg.Storage.BucketText,
		FileType:  "text/plain",
	})

	if err != nil {
		return fmt.Errorf("upload text: %w", err)
	}

	rec := transcribed.BrokerRecord{
		TextName: msg.FileName,
		TextID:   msg.FileID,
		TextSize: int64(len(resp.Text)),
		Language: resp.Language,
	}

	if err := kp.W.Write(ctx, rec); err != nil {
		return fmt.Errorf("kp.W.Write: %w", err)
	}

	return nil
}
