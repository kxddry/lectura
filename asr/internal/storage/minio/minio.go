package minio

import (
	"context"
	"fmt"
	"github.com/kxddry/lectura/asr/internal/config"
	"github.com/kxddry/lectura/asr/internal/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"net/url"
	"strings"
	"time"
)

type MinioClient struct {
	mclient *minio.Client
	url     string
}

func (m *MinioClient) GetLink() string {
	return m.url
}

func (m *MinioClient) Upload(ctx context.Context, fc storage.FileConfig) (url string, size int64, err error) {
	info, err := m.mclient.PutObject(ctx, fc.Bucket, fc.Filename, fc.File, fc.Size,
		minio.PutObjectOptions{ContentType: fc.MType})
	if err != nil {
		return "", 0, err
	}
	return fmt.Sprintf("%s%s", m.GetLink(), info.Location), info.Size, nil
}

func (m *MinioClient) PresignedGetURL(ctx context.Context, bucket, object string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := m.mclient.PresignedGetObject(ctx, bucket, object, expiry, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

func (m *MinioClient) UploadText(ctx context.Context, id string, text string) (string, error) {
	reader := io.NopCloser(strings.NewReader(text))
	size := int64(len(text))

	outUrl, _, err := m.Upload(ctx, storage.FileConfig{
		Filename: id + ".txt",
		FileID:   id,
		File:     reader,
		Size:     size,
		Bucket:   "texts",
		MType:    "text/plain",
	})
	if err != nil {
		return "", err
	}
	return outUrl, nil
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
