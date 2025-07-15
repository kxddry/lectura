package app

type App struct {
	Name   string `yaml:"name" env-required:"true"`
	Pubkey string `yaml:"pubkey" env-required:"true"`
}
