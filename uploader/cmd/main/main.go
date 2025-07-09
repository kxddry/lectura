package main

import (
	"context"
	"github.com/kxddry/lectura/shared/utils/config"
	"github.com/kxddry/lectura/shared/utils/logger"
	"github.com/kxddry/lectura/shared/utils/logger/handlers/sl"
	"github.com/kxddry/lectura/uploader/internal/broker/kafka"
	cc "github.com/kxddry/lectura/uploader/internal/config"
	"github.com/kxddry/lectura/uploader/internal/handlers"
	mini "github.com/kxddry/lectura/uploader/internal/storage/minio"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"os"
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

	mc, err := mini.New(cfg.Storage)
	if err != nil {
		log.Error("Error creating minio client", sl.Err(err))
		os.Exit(1)
	}
	bucket := cfg.Storage.BucketName
	err = mc.EnsureBucketExists(ctx, bucket)
	if err != nil {
		log.Error("Failed to ensure bucket exists", sl.Err(err))
		os.Exit(1)
	}

	if err = kafka.CheckAlive(cfg.Kafka.Brokers); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	w := kafka.New(&cfg.Kafka)

	e := echo.New()
	e.Use(middleware.BodyLimit("1G"))

	e.POST("/upload", handlers.UploadHandler(ctx, log, w, mc, bucket))
	log.Info("Server started at " + cfg.Server.Address)
	e.Logger.Fatal(e.Start(cfg.Server.Address))
}
