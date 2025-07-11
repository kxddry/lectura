package storage

import (
	"context"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"io"
)

type Storage interface {
	Upload(ctx context.Context, tc transcribed.BrokerRecord) (url string, size int64, err error)
	EnsureBucketExists(ctx context.Context, bucket string) error
	Download(ctx context.Context, FC uploaded.FileConfig) (io.ReadCloser, error)
	GetLink() string
}
