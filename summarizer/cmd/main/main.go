package main

import (
	"context"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/shared/utils/config"
	"github.com/kxddry/lectura/shared/utils/logger"
	"github.com/kxddry/lectura/shared/utils/logger/handlers/sl"
	"github.com/kxddry/lectura/summarizer/internal/broker/kafka"
	config2 "github.com/kxddry/lectura/summarizer/internal/config"
	"github.com/kxddry/lectura/summarizer/internal/handlers"
	"github.com/kxddry/lectura/summarizer/internal/llm"
	"github.com/kxddry/lectura/summarizer/internal/storage/minio"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const workerPoolSize = 100

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var cfg config2.Config
	config.MustParseConfig(&cfg)

	log := logger.SetupLogger(cfg.Env)
	log.Debug("debug enabled")

	if cfg.Summarizer.ApiKey == "" {
		log.Warn("API key is empty!")
	}

	mc, err := minio.New(cfg.Storage)
	if err != nil {
		log.Error("Error creating minio client", sl.Err(err))
		os.Exit(1)
	}
	if err = mc.EnsureBucketExists(ctx, cfg.Storage.BucketInput); err != nil {
		log.Error("Failed to ensure bucket for input exists", sl.Err(err))
		os.Exit(1)
	}
	if err = mc.EnsureBucketExists(ctx, cfg.Storage.BucketOutput); err != nil {
		log.Error("Failed to ensure bucket for output exists", sl.Err(err))
		os.Exit(1)
	}

	log.Debug("minio client created")

	if err = kafka.CheckAlive(cfg.Kafka.Brokers); err != nil {
		log.Error("CheckAlive failed", sl.Err(err))
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

	jobs := make(chan transcribed.BrokerRecord, 100)
	results := make(chan error, 100)

	for i := 0; i < workerPoolSize; i++ {

		// workers handling jobs
		go func(id int) {
			for msg := range jobs {
				err := handlers.Pipeline(ctx, &cfg, llm.OpenAI{&cfg}, mc, kp, msg)

				if err != nil {
					log.Error("error processing job", sl.Err(err))
				}
				results <- err
			}
		}(i)

	}
	log.Debug("worker pool created")

	go distributeJobs(ctx, log, kp, jobs)
	log.Debug("job handler started")

	go processResults(log, results)
	log.Debug("error handler started")

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("signal received, shutting down gracefully")
}

func distributeJobs(ctx context.Context, log *slog.Logger, kp handlers.KafkaPipeline, jobs chan<- transcribed.BrokerRecord) {
	msgCh := kp.InputCh
	errCh := kp.ErrCh
	for {
		// orchestrate
		select {
		case msg := <-msgCh:
			log.Debug("message sent to jobs", slog.String("msg", msg.TextID))
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
