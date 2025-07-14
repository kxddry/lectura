package main

import (
	"context"
	config2 "github.com/kxddry/lectura/asr/internal/config"
	"github.com/kxddry/lectura/asr/internal/handlers"

	// shared tools
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/kxddry/lectura/shared/utils/broker/kafka"
	"github.com/kxddry/lectura/shared/utils/config"
	"github.com/kxddry/lectura/shared/utils/logger"
	"github.com/kxddry/lectura/shared/utils/logger/handlers/sl"
	"github.com/kxddry/lectura/shared/utils/s3"

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
	var cfg config2.Config
	config.MustParseConfig(&cfg)

	// logger
	log := logger.SetupLogger(cfg.Env)
	log.Debug("debug enabled")

	// S3 client
	cli, err := s3.NewClient(cfg.S3Storage)
	if err != nil {
		log.Error("Error creating minio client", sl.Err(err))
		os.Exit(1)
	}

	log.Debug("minio client created")

	r := kafka.NewReader[uploaded.Record](cfg.Kafka.Read)
	if err = r.CheckAlive(); err != nil {
		log.Error("CheckAlive failed", sl.Err(err))
		os.Exit(1)
	}
	w := kafka.NewWriter[transcribed.Record](cfg.Kafka.Write)

	kp := kafka.NewPipeline(r, w)
	log.Debug("kafka clients created")

	// create a worker pool
	jobs := make(chan uploaded.Record, workerPoolSize*10)
	results := make(chan error, workerPoolSize*10)

	for i := 0; i < workerPoolSize; i++ {
		go func(id int) {
			for msg := range jobs {
				err := handlers.Pipeline(ctx, cfg, cli, kp, msg)
				if err != nil {
					log.Error("error processing job", sl.Err(err))
				}
				results <- err
			}
		}(i)

	}
	log.Debug("worker pool created")

	go distributeJobs(ctx, log, r, jobs)
	log.Debug("job handler started")

	go processResults(log, results)
	log.Debug("error handler started")

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("signal received, shutting down gracefully")
}

func distributeJobs[T uploaded.Record](ctx context.Context, log *slog.Logger, r kafka.Reader[T], jobs chan<- T) {
	msgCh, errCh := r.Messages(ctx)
	for {
		select {
		case msg := <-msgCh:
			jobs <- msg
		case err := <-errCh:
			log.Error("kafka reader", sl.Err(err))
			return
		case <-ctx.Done():
			log.Debug("distributor shutting down, ctx done")
			close(jobs)
			return
		}
	}
}

func processResults(log *slog.Logger, results <-chan error) {
	for err := range results {
		if err != nil {
			log.Error("error popped up", sl.Err(err))
		} else {
			log.Debug("job processed successfully")
		}
	}
}
