package config

import (
	"time"
)

type Config struct {
	Env     string  `yaml:"env" env-required:"true"`
	Storage Storage `yaml:"storage" env-required:"true"`
	Server  Server  `yaml:"server" env-required:"true"`
	Kafka   Kafka   `yaml:"kafka" env-required:"true"`
}
type Storage struct {
	Type            string `yaml:"type" env-required:"true"`
	Endpoint        string `yaml:"endpoint" env-required:"true"`
	AccessKeyID     string `yaml:"access_key_id" env-required:"true"`
	SecretAccessKey string `yaml:"secret_access_key" env-required:"true"`
	UseSSL          bool   `yaml:"ssl" env-default:"false"`
	BucketName      string `yaml:"bucket" env-default:"videos"`
}

type Server struct {
	Address     string        `yaml:"address" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-required:"true"`
}

type Kafka struct {
	Brokers         []string      `yaml:"brokers" env-required:"true"`
	Topic           string        `yaml:"topic" env-required:"true"`
	ClientID        string        `yaml:"client_id" env-required:"true"`
	Retries         int           `yaml:"retries" env-default:"5"`
	MaxMessageBytes int           `yaml:"max_message_bytes" env-default:"1048576"`
	Acks            string        `yaml:"acks" env-default:"all"`        // 0 | 1 | all
	Compression     string        `yaml:"compression" env-default:"lz4"` // lz4 | snappy | none | gzip | zstd
	Timeout         time.Duration `yaml:"timeout" env-default:"5s"`      // time.Duration
}
