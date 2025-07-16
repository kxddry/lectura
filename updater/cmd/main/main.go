package main

import (
	"context"
	"github.com/kxddry/lectura/shared/entities/summarized"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/kxddry/lectura/shared/utils/broker/kafka"
	"github.com/kxddry/lectura/shared/utils/config"
	"github.com/kxddry/lectura/shared/utils/logger"
	"github.com/kxddry/lectura/shared/utils/logger/handlers/sl"
	"github.com/kxddry/lectura/shared/utils/storage/postgres"
	cc "github.com/kxddry/lectura/updater/internal/config"
	"github.com/kxddry/lectura/updater/internal/handlers"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var cfg cc.Config
	config.MustParseConfig(&cfg)

	workerPoolSize := cfg.WorkerPoolSize
	multi := cfg.WorkerPoolMultiplier

	if len(cfg.KafkaTopics) != 3 {
		panic("3 topics required: uploaded, asm, sum")
	}

	log := logger.SetupLogger(cfg.Env)
	log.Debug("debug enabled")

	sql, err := postgres.New(cfg.Storage)
	if err != nil {
		panic(err)
	}
	log.Debug("sql created")

	jobs := make(chan any, workerPoolSize*multi)
	results := make(chan error, workerPoolSize*multi)

	for i := range workerPoolSize {
		go func(id int) {
			log.Debug("worker listening " + strconv.Itoa(id))
			for msg := range jobs {
				log.Info("msg received", msg)
				err = handlers.ProcessMessage(ctx, msg, sql)
				if err != nil {
					log.Error("error processing", sl.Err(err))
				}
				results <- err
			}
		}(i)
	}

	go processResults(ctx, log, results)
	run(ctx, &cfg, log, jobs)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("signal received, shutting down gracefully")
}

func run(ctx context.Context, cfg *cc.Config, log *slog.Logger, jobs chan<- any) {
	// reader 1
	cfg1 := cfg.Kafka
	cfg1.Topic = cfg.KafkaTopics[0]
	r1 := kafka.NewReader[uploaded.Record](cfg1)

	// reader 2
	cfg2 := cfg.Kafka
	cfg2.Topic = cfg.KafkaTopics[1]
	r2 := kafka.NewReader[transcribed.Record](cfg2)

	// reader 3
	cfg3 := cfg.Kafka
	cfg3.Topic = cfg.KafkaTopics[2]
	r3 := kafka.NewReader[summarized.Record](cfg3)

	go handleJobs(ctx, log, r1, jobs)
	go handleJobs(ctx, log, r2, jobs)
	go handleJobs(ctx, log, r3, jobs)
}

func handleJobs[T uploaded.Record | transcribed.Record | summarized.Record](
	ctx context.Context, log *slog.Logger, r kafka.Reader[T], jobs chan<- any) {
	msgCh, errCh := r.Messages(ctx)
	for {
		select {
		case msg := <-msgCh:
			jobs <- msg
		case err := <-errCh:
			log.Error("kafka reader", sl.Err(err))
			return
		case <-ctx.Done():
			log.Debug("ctx done")
			close(jobs)
			return
		}
	}
}

func processResults(ctx context.Context, log *slog.Logger, results <-chan error) {
	for {
		select {
		case err, ok := <-results:
			if !ok {
				return
			}
			if err != nil {
				log.Error("error popped up", sl.Err(err))
			}
		case <-ctx.Done():
			return
		}
	}

}
