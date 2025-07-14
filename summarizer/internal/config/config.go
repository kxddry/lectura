package config

import (
	kafka2 "github.com/kxddry/lectura/shared/entities/config/kafka"
)

type Config struct {
	Env        string     `yaml:"env" env-required:"true"`
	Summarizer Summarizer `yaml:"summarizer" env-required:"true"`
	Kafka      Kafka      `yaml:"kafka" env-required:"true"`
}

type Summarizer struct {
	BaseUrl string `yaml:"base_url" env-required:"true"`
	Model   string `yaml:"model" env-required:"true"`
	Prompt  string `yaml:"prompt" env-required:"true"`
	ApiKey  string `env:"OPENAI_API_KEY" env-required:"false"`
}

type Kafka struct {
	Reader kafka2.ReaderConfig `yaml:"reader" env-required:"true"`
	Writer kafka2.WriterConfig `yaml:"writer" env-required:"true"`
}
