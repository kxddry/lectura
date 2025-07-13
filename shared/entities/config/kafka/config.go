package kafka

import "time"

type WriterConfig struct {
	Brokers         []string      `yaml:"brokers" env-required:"true"`
	Topic           string        `yaml:"topic" env-required:"true"`
	ClientID        string        `yaml:"client_id" env-required:"true"`
	Retries         int           `yaml:"retries" env-default:"5"`
	MaxMessageBytes int           `yaml:"max_message_bytes" env-default:"1048576"`
	Acks            string        `yaml:"acks" env-default:"all"`        // 0 | 1 | all
	Compression     string        `yaml:"compression" env-default:"lz4"` // lz4 | snappy | none | gzip | zstd
	Timeout         time.Duration `yaml:"timeout" env-default:"5s"`      // time.Duration
}

type ReaderConfig struct {
	Brokers        []string      `yaml:"brokers" env-required:"true"`
	Topic          string        `yaml:"topic" env-required:"true"`
	GroupID        string        `yaml:"group_id" env-required:"true"`
	MinBytes       int           `yaml:"min_bytes" env-default:"1"`         // min fetch bytes
	MaxBytes       int           `yaml:"max_bytes" env-default:"1048576"`   // 1MB
	CommitInterval time.Duration `yaml:"commit_interval" env-default:"1s"`  // time.Duration, e.g. 1s
	StartOffset    string        `yaml:"start_offset" env-default:"latest"` // earliest | latest
}
