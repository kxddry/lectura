package minio

import (
	"context"
	"github.com/kxddry/lectura/uploader/internal/config"
	"github.com/kxddry/lectura/uploader/internal/entities"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	mclient *minio.Client
}

func (m *MinioClient) Upload(ctx context.Context, file entities.File, bucket string) error {
	_, err := m.mclient.PutObject(ctx, bucket, file.UUID+file.Extension, file.Data, file.Size, minio.PutObjectOptions{ContentType: file.Type})
	return err
}

func EnsureBucketExists(m *MinioClient, ctx context.Context, bucket string) error {
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

func New(cfg config.S3Storage) (*MinioClient, error) {
	mclient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &MinioClient{mclient}, nil
}
