package storage

import (
	"context"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"io"
)

type FileConfig struct {
	Filename string
	FileID   string
	File     io.ReadCloser
	Size     int64
	Bucket   string
	MType    string
}

type Storage interface {
	Upload(ctx context.Context, tc transcribed.BrokerRecord) (url string, size int64, err error)
	EnsureBucketExists(ctx context.Context, bucket string) error
	GetLink() string
}
