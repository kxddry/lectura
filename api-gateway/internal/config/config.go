package config

import (
	"github.com/kxddry/lectura/shared/entities/auth"
	"github.com/kxddry/lectura/shared/entities/config/app"
	"github.com/kxddry/lectura/shared/entities/config/db"
	"github.com/kxddry/lectura/shared/entities/config/s3"
	"github.com/kxddry/lectura/shared/entities/config/services"
	"time"
)

type Config struct {
	Services    `yaml:"services" env-required:"true"`
	Expiry      time.Duration         `yaml:"expiry" env-default:"1h"`
	Env         string                `yaml:"env" env-required:"true"`
	Server      services.Server       `yaml:"server" env-required:"true"`
	RateLimit   int                   `yaml:"rate_limit" env:"RATE_LIMIT" env-default:"100"`
	App         app.App               `yaml:"app" env-required:"true"`
	PubkeyPath  string                `yaml:"pubkey_path" env:"PUBKEY_PATH" env-required:"true"`
	PrivkeyPath string                `yaml:"privkey_path" env:"PRIVKEY_PATH"`
	PublicKeys  []auth.PublicKeyEntry `yaml:"public_keys"`
	Storage     db.StorageConfig      `yaml:"storage" env-required:"true"`
	S3Storage   s3.StorageConfig      `yaml:"s3storage" env-required:"true"`
}

type Services struct {
	Auth services.SSO `yaml:"auth" env-required:"true"`
}
