package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kxddry/lectura/asr/internal/config"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"io"
	"net/http"
	"time"
)

type Uploader interface {
	UploadText(ctx context.Context, id string, text string) (string, error)
}

type URLPresigner interface {
	PresignedGetURL(ctx context.Context, bucket, object string, expiry time.Duration) (string, error)
}

type URLUploader interface {
	Uploader
	URLPresigner
}

type Writer interface {
	Write(ctx context.Context, record transcribed.BrokerRecord) error
}

type KafkaPipeline struct {
	InputCh <-chan uploaded.BrokerRecord
	ErrCh   <-chan error

	W Writer
}

func callWhisperAPI(apiUrl string, tr transcribed.TranscribeRequest) (*transcribed.TranscribeResponse, error) {
	reqBody, _ := json.Marshal(tr)
	resp, err := http.Post(apiUrl, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var respObj transcribed.TranscribeResponse

	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return nil, err
	}

	return &respObj, nil
}

func Pipeline(ctx context.Context, cfg *config.Config, up URLUploader, kp KafkaPipeline, msg uploaded.BrokerRecord) error {
	objectKey := msg.FileName
	signedURL, err := up.PresignedGetURL(ctx, cfg.Storage.BucketInput, objectKey, cfg.Storage.TTL)
	if err != nil {
		return fmt.Errorf("presign url: %w", err)
	}

	tr := transcribed.TranscribeRequest{
		ID:       msg.FileID,
		AudioURL: signedURL,
	}

	resp, err := callWhisperAPI(cfg.WhisperAPI, tr)
	if err != nil {
		return fmt.Errorf("callWhisperAPI: %w", err)
	}

	URL, err := up.UploadText(ctx, resp.ID, resp.Text)
	if err != nil {
		return fmt.Errorf("upload text: %w", err)
	}
	rec := transcribed.BrokerRecord{
		ID:       resp.ID,
		TextUrl:  URL,
		Duration: resp.Duration,
		Language: resp.Language,
	}

	if err := kp.W.Write(ctx, rec); err != nil {
		return fmt.Errorf("kp.W.Write: %w", err)
	}

	return nil
}
