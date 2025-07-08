package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log/slog"
	"os"
	"uploader/config"
	"uploader/internal/lib/logger"

	"uploader/internal/lib/handlers"
)

func main() {
	ctx := context.Background()
	cfg := config.MustLoad()
	if cfg.Storage.Type != "minio" {
		panic("Invalid storage type. Currently supported: minio.")
	}

	log := logger.SetupLogger(cfg.Env)
	mc, err := newMinioClient(cfg)

	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	bucketName := cfg.Storage.BucketName

	ensureBucketExists(ctx, log, mc, bucketName)

	e := echo.New()
	e.Use(middleware.BodyLimit("2G"))

	e.POST("/upload", handlers.UploadHandler(ctx, log, mc, bucketName))
	log.Info("Server started at " + cfg.Server.Address)
	e.Logger.Fatal(e.Start(cfg.Server.Address))
}

func ensureBucketExists(ctx context.Context, log *slog.Logger, mc *minio.Client, bucketName string) {
	err := mc.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := mc.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Info("Bucket" + bucketName + "already exists")
		} else {
			log.Error(err.Error())
			os.Exit(1)
		}
	} else {
		log.Info("Successfully created bucket " + bucketName)
	}
}

func newMinioClient(cfg *config.Config) (*minio.Client, error) {
	return minio.New(
		cfg.Storage.Endpoint, &minio.Options{
			Creds: credentials.NewStaticV4(
				cfg.Storage.AccessKeyID, cfg.Storage.SecretAccessKey, ""),
			Secure: cfg.Storage.UseSSL,
		},
	)
}
