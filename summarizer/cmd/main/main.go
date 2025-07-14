package main

import (
	"context"
	"github.com/kxddry/lectura/shared/entities/summarized"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/shared/utils/broker/kafka"
	"github.com/kxddry/lectura/shared/utils/config"
	"github.com/kxddry/lectura/shared/utils/logger"
	"github.com/kxddry/lectura/shared/utils/logger/handlers/sl"
	config2 "github.com/kxddry/lectura/summarizer/internal/config"
	"github.com/kxddry/lectura/summarizer/internal/handlers"
	"github.com/kxddry/lectura/summarizer/internal/llm"
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

	r := kafka.NewReader[transcribed.Record](cfg.Kafka.Reader)
	w := kafka.NewWriter[summarized.Record](cfg.Kafka.Writer)

	kp := kafka.NewPipeline(r, w)

	// create a worker pool

	jobs := make(chan transcribed.Record, 100)
	results := make(chan error, 100)

	for i := 0; i < workerPoolSize; i++ {

		// workers handling jobs
		go func(id int) {
			for msg := range jobs {
				err := handlers.Pipeline(ctx, llm.OpenAI{&cfg}, kp, msg)
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

func distributeJobs[R transcribed.Record, W summarized.Record](ctx context.Context, log *slog.Logger, kp kafka.Pipeline[R, W], jobs chan<- R) {
	msgCh, errCh := kp.R.Messages(ctx)
	for {
		// orchestrate
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
