package services

import "time"

type SSO struct {
	Address  string        `yaml:"address" env-required:"true"`
	Timeout  time.Duration `yaml:"timeout" env-default:"5s"`
	Retries  int           `yaml:"retries" env-default:"5"`
	Insecure bool          `yaml:"insecure" env-default:"true"`
}
