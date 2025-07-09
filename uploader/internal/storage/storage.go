package storage

import (
	"context"
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
	Upload(ctx context.Context, fc FileConfig) (url string, size int64, err error)
	EnsureBucketExists(ctx context.Context, bucket string) error
	GetLink() string
}
