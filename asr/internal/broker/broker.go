package broker

import (
	"context"
	"errors"
)

type BrokerRecord struct {
	FileName   string `json:"file_name"`
	FileID     string `json:"file_id"`
	FileType   string `json:"file_type"`
	FileSize   int64  `json:"file_size"`
	WavSize    int64  `json:"wav_size"`
	UploadedAt int64  `json:"uploaded_at"`
	S3URL      string `json:"s3_url"`
	WavURL     string `json:"wav_url"`
	// UserID     string `json:"user_id"` TODO: implement auth
}

type Broker interface {
	CheckAlive() error
}

type Reader interface {
	Read(context.Context) (BrokerRecord, error)
}

type Writer interface {
	Write(context.Context, BrokerRecord) error
	Broker
}

var (
	ErrUninitializedBroker = errors.New("broker not initialized")
)
