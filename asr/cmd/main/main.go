package main

import (
	"context"
	"github.com/kxddry/lectura/asr/internal/broker/kafka"
	cc "github.com/kxddry/lectura/asr/internal/config"
	"github.com/kxddry/lectura/asr/internal/handlers"
	"github.com/kxddry/lectura/asr/internal/storage/minio"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/kxddry/lectura/shared/utils/config"
	"github.com/kxddry/lectura/shared/utils/logger"
	"github.com/kxddry/lectura/shared/utils/logger/handlers/sl"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const workerPoolSize = 10

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config
	var cfg cc.Config
	config.MustParseConfig(&cfg)

	// logger
	log := logger.SetupLogger(cfg.Env)
	log.Debug("debug enabled")

	// S3 client
	mc, err := minio.New(cfg.Storage)
	if err != nil {
		log.Error("Error creating minio client", sl.Err(err))
		os.Exit(1)
	}
	if err = mc.EnsureBucketExists(ctx, cfg.Storage.BucketInput); err != nil {
		log.Error("Failed to ensure bucket for input exists", sl.Err(err))
		os.Exit(1)
	}
	if err = mc.EnsureBucketExists(ctx, cfg.Storage.BucketText); err != nil {
		log.Error("Failed to ensure bucket for text exists", sl.Err(err))
		os.Exit(1)
	}

	log.Debug("minio client created")

	// kafka
	if err = kafka.CheckAlive(cfg.Kafka.Brokers); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// read uploaded files
	r := kafka.NewReader(&cfg.Kafka)
	msgCh, errCh := r.Messages(ctx)
	// and send out texts
	kp := handlers.KafkaPipeline{
		msgCh, errCh, kafka.NewWriter(&cfg.Kafka),
	}
	log.Debug("kafka clients created")

	// create a worker pool

	jobs := make(chan uploaded.BrokerRecord, 100)
	results := make(chan error, 100)
	for i := 0; i < workerPoolSize; i++ {
		go func(id int) {
			// process incoming jobs
			for msg := range jobs {
				err := handlers.Pipeline(ctx, &cfg, mc, kp, msg)

				if err != nil {
					log.Error("error processing job", sl.Err(err))
				} else {
					log.Debug("job processed successfully")
				}
				results <- err
			}
		}(i)
	}
	log.Debug("worker pool created")

	go func() {
		for {
			// orchestrate
			select {
			case msg := <-msgCh:
				log.Debug("message sent to jobs", slog.String("file name", msg.FileName),
					slog.String("file id", msg.FileID),
					slog.String("file type", msg.FileType))
				jobs <- msg
			case err := <-errCh:
				log.Error("kafka reader", sl.Err(err))
				return
			case <-ctx.Done():
				log.Debug("orchestrator shutting down, ctx done")
				close(jobs)
				return
			}
		}
	}()
	log.Debug("orchestrator started")

	go func() {
		// process errors
		for err := range results {
			if err != nil {
				log.Error("error popped up", sl.Err(err))
			}
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("signal received, shutting down gracefully")
}
