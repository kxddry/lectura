package config

import "time"

type Config struct {
	Env        string  `yaml:"env" env-required:"true"`
	WhisperAPI string  `yaml:"whisper_api" env-required:"true"`
	Storage    Storage `yaml:"storage" env-required:"true"`
	Kafka      Kafka   `yaml:"kafka" env-required:"true"`
}
type Storage struct {
	Type            string        `yaml:"type" env-required:"true"`
	Endpoint        string        `yaml:"endpoint" env-required:"true"`
	AccessKeyID     string        `yaml:"access_key_id" env-required:"true"`
	SecretAccessKey string        `yaml:"secret_access_key" env-required:"true"`
	UseSSL          bool          `yaml:"ssl" env-default:"false"`
	BucketInput     string        `yaml:"bucket_input" env-default:"input"`
	BucketText      string        `yaml:"bucket_text" env-default:"text"`
	TTL             time.Duration `yaml:"ttl" env-default:"15m"` // for presigned urls
}

type Kafka struct {
	Brokers []string `yaml:"brokers" env-required:"true"`
	Read    Read     `yaml:"read" env-required:"true"`
	Write   Write    `yaml:"write" env-required:"true"`
}

type Read struct {
	Topic          string        `yaml:"topic" env-required:"true"`
	GroupID        string        `yaml:"group_id" env-required:"true"`
	MinBytes       int           `yaml:"min_bytes" env-default:"1"`         // min fetch bytes
	MaxBytes       int           `yaml:"max_bytes" env-default:"1048576"`   // 1MB
	CommitInterval time.Duration `yaml:"commit_interval" env-default:"1s"`  // time.Duration, e.g. 1s
	StartOffset    string        `yaml:"start_offset" env-default:"latest"` // earliest | latest
}

type Write struct {
	Topic           string        `yaml:"topic" env-required:"true"`
	ClientID        string        `yaml:"client_id" env-required:"true"`
	Retries         int           `yaml:"retries" env-default:"5"`
	MaxMessageBytes int           `yaml:"max_message_bytes" env-default:"1048576"`
	Acks            string        `yaml:"acks" env-default:"all"`        // 0 | 1 | all
	Compression     string        `yaml:"compression" env-default:"lz4"` // lz4 | snappy | none | gzip | zstd
	Timeout         time.Duration `yaml:"timeout" env-default:"5s"`      // time.Duration
}
