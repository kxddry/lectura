package s3

import (
	"context"
	"github.com/kxddry/lectura/shared/entities/config/s3"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"net/url"
	"time"
)

type S3Client struct {
	internal  *minio.Client
	public    *minio.Client
	publicURL *url.URL
}

func (s3 S3Client) GetPresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)

	presignedURL, err := s3.public.PresignedGetObject(ctx, bucket, objectName, expiry, reqParams)
	if err != nil {
		return "", err
	}

	parsed := *presignedURL
	parsed.Host = s3.publicURL.Host
	parsed.Scheme = s3.publicURL.Scheme

	return parsed.String(), nil
}

func NewClient(config s3.StorageConfig) (S3Client, error) {
	internal, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.Secret, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return S3Client{}, err
	}

	u, err := url.Parse(config.PublicURL)
	if err != nil {
		return S3Client{}, err
	}

	secure := u.Scheme == "https"

	publicClient, err := minio.New(u.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.Secret, ""),
		Secure: secure,
	})
	if err != nil {
		return S3Client{}, err
	}
	return S3Client{internal, publicClient, u}, nil
}

func (s3 S3Client) Download(ctx context.Context, bucket string, key string) (io.ReadCloser, error) {
	return s3.internal.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
}

func (s3 S3Client) Upload(ctx context.Context, bucket string, file uploaded.File) error {
	_, err := s3.internal.PutObject(ctx, bucket, file.FullName(), file.Data(), file.Size(), minio.PutObjectOptions{
		ContentType: file.MimeType(),
	})
	return err
}

func (s3 S3Client) Delete(ctx context.Context, bucket string, key string) error {
	return s3.internal.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

func EnsureBucketExists(ctx context.Context, s3 S3Client, bucket string) error {
	err := s3.internal.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := s3.internal.BucketExists(ctx, bucket)
		if errBucketExists == nil && exists {
			return nil
		}
		return err
	}
	return nil
}
