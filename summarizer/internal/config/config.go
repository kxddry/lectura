package config

import "time"

type Config struct {
	Env        string     `yaml:"env" env-required:"true"`
	Summarizer Summarizer `yaml:"summarizer" env-required:"true"`
	Storage    Storage    `yaml:"storage" env-required:"true"`
	Kafka      Kafka      `yaml:"kafka" env-required:"true"`
}

type Summarizer struct {
	BaseUrl string `yaml:"base_url" env-required:"true"`
	Model   string `yaml:"model" env-required:"true"`
	Prompt  string `yaml:"prompt" env-required:"true"`
	ApiKey  string `env:"OPENAI_API_KEY" env-required:"false"`
}

type Storage struct {
	Type         string `yaml:"type" env-required:"true"`
	Endpoint     string `yaml:"endpoint" env-required:"true"`
	AccessKey    string `yaml:"access_key_id" env-required:"true"`
	SecretKey    string `yaml:"secret_access_key" env-required:"true"`
	UseSSL       bool   `yaml:"ssl" env-default:"false"`
	BucketInput  string `yaml:"bucket_input" env-required:"true"`
	BucketOutput string `yaml:"bucket_output" env-required:"true"`
}

type Kafka struct {
	Brokers []string `yaml:"brokers" env-required:"true"`
	Read    struct {
		Topic          string        `yaml:"topic" env-required:"true"`
		GroupID        string        `yaml:"group_id" env-required:"true"`
		MinBytes       int           `yaml:"min_bytes" env-default:"1"`
		MaxBytes       int           `yaml:"max_bytes" env-default:"1048576"`
		CommitInterval time.Duration `yaml:"commit_interval" env-default:"3s"`
		StartOffset    string        `yaml:"start_offset" env-default:"latest"`
	} `yaml:"read" env-required:"true"`
	Write struct {
		Topic           string        `yaml:"topic" env-required:"true"`
		ClientID        string        `yaml:"client_id" env-required:"true"`
		Retries         int           `yaml:"retries" env-default:"5"`
		MaxMessageBytes int           `yaml:"max_message_bytes" env-default:"1048576"`
		Timeout         time.Duration `yaml:"timeout" env-default:"5s"`
		Acks            string        `yaml:"acks" env-default:"all"` // 0 | 1 | all
		Compression     string        `yaml:"compression" env-default:"lz4"`
	} `yaml:"write" env-required:"true"`
}
