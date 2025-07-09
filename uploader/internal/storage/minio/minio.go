package minio

import (
	"context"
	"fmt"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/kxddry/lectura/uploader/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	mclient *minio.Client
	url     string
}

func (m *MinioClient) GetLink() string {
	return m.url
}

func (m *MinioClient) Upload(ctx context.Context, fc uploaded.FileConfig) (url string, size int64, err error) {
	info, err := m.mclient.PutObject(ctx, fc.Bucket, fc.Filename, fc.File, fc.Size,
		minio.PutObjectOptions{ContentType: fc.MType})
	if err != nil {
		return "", 0, err
	}
	return fmt.Sprintf("%s%s", m.url, info.Location), info.Size, nil
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
		url:     mclient.EndpointURL().String(),
	}, nil
}
