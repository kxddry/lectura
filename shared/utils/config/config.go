package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

// MustParseConfig panics if the config wasn't loaded.
// Usage: create a pointer to your config and input it here.
// Example: MustParseConfig(&cfg)
func MustParseConfig(cfg any) {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file doesn't exist " + path)
	}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		panic(err)
	}
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
