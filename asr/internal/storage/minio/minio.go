package minio

import (
	"context"
	"github.com/kxddry/lectura/asr/internal/config"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
)

type MinioClient struct {
	mclient *minio.Client
}

func (m *MinioClient) Upload(ctx context.Context, fc uploaded.FileConfig) error {
	_, err := m.mclient.PutObject(ctx, fc.Bucket, fc.FileID+fc.Extension, fc.File, fc.FileSize,
		minio.PutObjectOptions{ContentType: fc.FileType})
	return err
}

func (m *MinioClient) Download(ctx context.Context, fc uploaded.FileConfig) (io.ReadCloser, error) {
	return m.mclient.GetObject(ctx, fc.Bucket, fc.FileID+fc.Extension, minio.GetObjectOptions{})
}

func (m *MinioClient) EnsureBucketExists(ctx context.Context, bucket string) error {
	err := m.mclient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := m.mclient.BucketExists(ctx, bucket)
		if errBucketExists == nil && exists {
			return nil
		}
		return err
	}
	return nil
}

func New(cfg config.Storage) (*MinioClient, error) {
	mclient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &MinioClient{
		mclient: mclient,
	}, nil
}
