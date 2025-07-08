package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env     string  `yaml:"env" env-required:"true"`
	Storage Storage `yaml:"storage" env-required:"true"`
	Server  Server  `yaml:"server" env-required:"true"`
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

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}
	return MustLoadByPath(path)
}

func MustLoadByPath(path string) *Config {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file doesn't exist " + path)
	}
	var res Config
	if err := cleanenv.ReadConfig(path, &res); err != nil {
		panic(err)
	}
	return &res
}

// fetchConfigPath parses the config path from flags or env and returns it.
// It prioritizes flags over env.
// Default return value: empty string
func fetchConfigPath() string {
	var res string
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()
	if res != "" {
		return res
	}
	env := os.Getenv("CONFIG_PATH")
	return env
}
