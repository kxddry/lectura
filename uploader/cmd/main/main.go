package main

import (
	"context"
	"github.com/kxddry/lectura/shared/utils/config"
	"github.com/kxddry/lectura/shared/utils/logger"
	"github.com/kxddry/lectura/shared/utils/logger/handlers/sl"
	middleware2 "github.com/kxddry/lectura/shared/utils/middleware"
	"github.com/kxddry/lectura/uploader/internal/broker/kafka"
	cc "github.com/kxddry/lectura/uploader/internal/config"
	"github.com/kxddry/lectura/uploader/internal/handlers"
	mini "github.com/kxddry/lectura/uploader/internal/storage/minio"
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
	if cfg.S3Storage.Type != "minio" {
		panic("Invalid storage type. Currently supported: minio.")
	}

	secret := []byte(cfg.AppSecret)

	// init logger
	log := logger.SetupLogger(cfg.Env)
	log.Debug("debug enabled")

	// init S3 client
	mc, err := mini.New(cfg.S3Storage)
	if err != nil {
		log.Error("Error creating minio client", sl.Err(err))
		os.Exit(1)
	}

	err = mini.EnsureBucketExists(mc, ctx, cfg.S3Storage.BucketName)
	if err != nil {
		log.Error("Failed to ensure bucket exists", sl.Err(err))
		os.Exit(1)
	}

	// init kafka writer
	if err = kafka.CheckAlive(cfg.Kafka.Brokers); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	w := kafka.New(&cfg.Kafka)

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
	e.Use(middleware.BodyLimit("1G"))
	e.Use(middleware2.JWTFromCookie(secret))

	e.POST("/api/upload", handlers.UploadHandler(ctx, log, w, mc, cfg))
	log.Info("Server started at " + cfg.Server.Address)
	e.Logger.Fatal(e.Start(cfg.Server.Address))
}
