package config

import (
	"github.com/kxddry/lectura/shared/entities/auth"
	"github.com/kxddry/lectura/shared/entities/config/app"
	"github.com/kxddry/lectura/shared/entities/config/services"
)

type Config struct {
	Services `yaml:"services" env-required:"true"`

	Env         string                `yaml:"env" env-required:"true"`
	Server      services.Server       `yaml:"server" env-required:"true"`
	RateLimit   int                   `yaml:"rate_limit" env:"RATE_LIMIT" env-default:"100"`
	App         app.App               `yaml:"app" env-required:"true"`
	PubkeyPath  string                `yaml:"pubkey_path" env:"PUBKEY_PATH" env-required:"true"`
	PrivkeyPath string                `yaml:"privkey_path" env:"PRIVKEY_PATH"`
	PublicKeys  []auth.PublicKeyEntry `yaml:"public_keys"`
}

type Services struct {
	Auth services.SSO `yaml:"auth" env-required:"true"`
}
