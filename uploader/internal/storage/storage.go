package storage

import (
	"context"
	"github.com/kxddry/lectura/shared/entities/uploaded"
)

type Storage interface {
	Upload(ctx context.Context, fc uploaded.FileConfig) (url string, size int64, err error)
	EnsureBucketExists(ctx context.Context, bucket string) error
	GetLink() string
}
