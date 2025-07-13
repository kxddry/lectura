package s3

import (
	"context"
	"github.com/kxddry/lectura/shared/entities/config/s3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
)

type File interface {
	FullName() string
	Data() io.Reader
	Size() int64
	MimeType() string
}

type S3Client struct {
	client *minio.Client
}

func NewClient(config s3.StorageConfig) (S3Client, error) {
	cli, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.Secret, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return S3Client{}, err
	}
	return S3Client{client: cli}, nil
}

func (s3 S3Client) Download(ctx context.Context, bucket string, key string) (io.ReadCloser, error) {
	return s3.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
}

func (s3 S3Client) Upload(ctx context.Context, bucket string, file File) error {
	_, err := s3.client.PutObject(ctx, bucket, file.FullName(), file.Data(), file.Size(), minio.PutObjectOptions{
		ContentType: file.MimeType(),
	})
	return err
}

func (s3 S3Client) Delete(ctx context.Context, bucket string, key string) error {
	return s3.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

func (s3 S3Client) EnsureBucketExists(ctx context.Context, bucket string) error {
	err := s3.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := s3.client.BucketExists(ctx, bucket)
		if errBucketExists == nil && exists {
			return nil
		}
		return err
	}
	return nil
}
