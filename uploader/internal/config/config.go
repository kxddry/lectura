package config

import (
	"github.com/kxddry/lectura/shared/entities/auth"
	"github.com/kxddry/lectura/shared/entities/config/kafka"
	"github.com/kxddry/lectura/shared/entities/config/s3"
	"github.com/kxddry/lectura/shared/entities/config/services"
)

type Config struct {
	Env         string                `yaml:"env" env-required:"true"`
	S3Storage   s3.StorageConfig      `yaml:"s3storage" env-required:"true"`
	Server      services.Server       `yaml:"server" env-required:"true"`
	Kafka       kafka.WriterConfig    `yaml:"writer" env-required:"true"`
	Clients     Clients               `yaml:"clients" env-required:"true"`
	PubkeyPath  string                `yaml:"pubkey_path" env-required:"true"`
	PrivkeyPath string                `yaml:"privkey_path"`
	PublicKeys  []auth.PublicKeyEntry `yaml:"public_keys" env-required:"true"`
}

type Clients struct {
	SSO services.SSO `yaml:"sso" env-required:"true"`
}
