package config

import (
	"github.com/kxddry/lectura/shared/entities/config/kafka"
	"github.com/kxddry/lectura/shared/entities/config/s3"
)

type Config struct {
	Env        string           `yaml:"env" env-required:"true"`
	WhisperAPI string           `yaml:"whisper_api" env-required:"true"`
	S3Storage  s3.StorageConfig `yaml:"s3storage" env-required:"true"`
	Kafka      Kafka            `yaml:"kafka" env-required:"true"`
}

type Kafka struct {
	Brokers []string           `yaml:"brokers" env-required:"true"`
	Read    kafka.ReaderConfig `yaml:"read" env-required:"true"`
	Write   kafka.WriterConfig `yaml:"write" env-required:"true"`
}
