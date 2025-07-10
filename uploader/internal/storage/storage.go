package storage

import (
	"context"
	"github.com/kxddry/lectura/shared/entities/uploaded"
)

type Storage interface {
	Upload(ctx context.Context, fc uploaded.FileConfig) error
	EnsureBucketExists(ctx context.Context, bucket string) error
}
