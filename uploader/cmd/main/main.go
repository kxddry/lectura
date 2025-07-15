package main

import (
	"context"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/kxddry/lectura/shared/utils/broker/kafka"
	"github.com/kxddry/lectura/shared/utils/config"
	"github.com/kxddry/lectura/shared/utils/ed25519"
	"github.com/kxddry/lectura/shared/utils/logger"
	"github.com/kxddry/lectura/shared/utils/logger/handlers/sl"
	middleware2 "github.com/kxddry/lectura/shared/utils/middleware"
	"github.com/kxddry/lectura/shared/utils/s3"
	cc "github.com/kxddry/lectura/uploader/internal/config"
	"github.com/kxddry/lectura/uploader/internal/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"os"
)

func main() {
	// configure context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// parse config
	var cfg cc.Config
	config.MustParseConfig(&cfg)

	// init logger
	log := logger.SetupLogger(cfg.Env)
	log.Debug("debug enabled")

	bucket := os.Getenv("BUCKET")
	if bucket == "" {
		log.Warn("Environment variable BUCKET is not set. Set BUCKET=input.")
		bucket = "input"
	}

	// init S3 client
	s3Client, err := s3.NewClient(cfg.S3Storage)
	if err != nil {
		log.Error("Error creating minio client", sl.Err(err))
		os.Exit(1)
	}

	if err = s3.EnsureBucketExists(ctx, s3Client, bucket); err != nil {
		log.Error("Failed to ensure bucket exists", sl.Err(err))
		os.Exit(1)
	}

	pubKeyMap, err := ed25519.LoadPublicKeys(cfg.PublicKeys)
	if err != nil {
		log.Error("Error loading public keys", sl.Err(err))
		os.Exit(1)
	}

	// init kafka writer
	w := kafka.NewWriter[uploaded.Record](cfg.Kafka)

	// init router
	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))
	e.Use(middleware.BodyLimit("1GB"))
	e.Use(middleware2.JWTMiddleware(middleware2.JWTMiddlewareConfig{
		PublicKeys: pubKeyMap,
		CookieName: "access_token",
	}))

	e.POST("/api/v1/upload", handlers.UploadHandler(ctx, log, w, s3Client, bucket))
	log.Info("Server started at " + cfg.Server.Address)
	e.Logger.Fatal(e.Start(cfg.Server.Address))
}
