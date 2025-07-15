package config

import (
	"github.com/kxddry/lectura/shared/entities/config/db"
	"github.com/kxddry/lectura/shared/entities/config/kafka"
)

type Config struct {
	Env                  string             `yaml:"env" env-default:"prod"`
	Kafka                kafka.ReaderConfig `yaml:"kafka" env-required:"true"`
	Storage              db.StorageConfig   `yaml:"storage" env-required:"true"`
	KafkaTopics          []string           `yaml:"kafka_topics" env-required:"true"`
	WorkerPoolSize       int                `yaml:"worker_pool_size" env-default:"2"`
	WorkerPoolMultiplier int                `yaml:"worker_pool_multiplier" env-default:"2"`
}
