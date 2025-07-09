package main

import (
	"context"
	"github.com/kxddry/go-utils/pkg/config"
	"github.com/kxddry/go-utils/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
	"log/slog"
	"os"
	"time"
	cc "uploader/config"
	"uploader/internal/handlers"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var cfg cc.Config

	config.MustParseConfig(&cfg)
	if cfg.Storage.Type != "minio" {
		panic("Invalid storage type. Currently supported: minio.")
	}

	log := logger.SetupLogger(cfg.Env)
	log.Debug("debug enabled")

	mc, err := newMinioClient(&cfg)

	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	if err = checkKafkaAlive(cfg.Kafka.Brokers); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	w := newKafkaWriter(&cfg)
	bucketName := cfg.Storage.BucketName

	ensureBucketExists(ctx, log, mc, bucketName)

	e := echo.New()
	e.Use(middleware.BodyLimit("1G"))

	e.POST("/upload", handlers.UploadHandler(ctx, log, w, mc, bucketName))
	log.Info("Server started at " + cfg.Server.Address)
	e.Logger.Fatal(e.Start(cfg.Server.Address))
}

func newKafkaWriter(cfg *cc.Config) *kafka.Writer {

	var compression kafka.Compression
	switch cfg.Kafka.Compression {
	case "gzip":
		compression = kafka.Gzip
	case "snappy":
		compression = kafka.Snappy
	case "lz4":
		compression = kafka.Lz4
	case "zstd":
		compression = kafka.Zstd
	default:
		compression = compress.None
	}

	var requiredAcks kafka.RequiredAcks
	switch cfg.Kafka.Acks {
	case "0":
		requiredAcks = kafka.RequireNone
	case "1":
		requiredAcks = kafka.RequireOne
	case "all":
		requiredAcks = kafka.RequireAll
	default:
		requiredAcks = kafka.RequireAll
	}

	return &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Kafka.Brokers...),
		Topic:                  cfg.Kafka.Topic,
		Balancer:               &kafka.RoundRobin{},
		MaxAttempts:            cfg.Kafka.Retries,
		RequiredAcks:           requiredAcks,
		Async:                  true,
		Compression:            compression,
		WriteTimeout:           cfg.Kafka.Timeout,
		AllowAutoTopicCreation: false,
	}
}

func checkKafkaAlive(brokers []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Controller()
	return err
}

func ensureBucketExists(ctx context.Context, log *slog.Logger, mc *minio.Client, bucketName string) {
	err := mc.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := mc.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Info("Bucket " + bucketName + " already exists")
		} else {
			log.Error(err.Error())
			os.Exit(1)
		}
	} else {
		log.Info("Successfully created bucket " + bucketName)
	}
}

func newMinioClient(cfg *cc.Config) (*minio.Client, error) {
	return minio.New(
		cfg.Storage.Endpoint, &minio.Options{
			Creds: credentials.NewStaticV4(
				cfg.Storage.AccessKeyID, cfg.Storage.SecretAccessKey, ""),
			Secure: cfg.Storage.UseSSL,
		},
	)
}
