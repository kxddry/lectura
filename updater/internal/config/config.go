package config

import (
	"github.com/kxddry/lectura/shared/entities/config/db"
	"github.com/kxddry/lectura/shared/entities/config/kafka"
)

type Config struct {
	Env        string             `yaml:"env" env-default:"prod"`
	Kafka      kafka.ReaderConfig `yaml:"kafka" env-required:"true"`
	Storage    db.StorageConfig   `yaml:"storage" env-required:"true"`
	Migrations struct {
		Path string `yaml:"path" env-required:"true"`
	} `yaml:"migrations" env-required:"true"`
	KafkaTopics []string `yaml:"kafka_topics" env-required:"true"`
}
