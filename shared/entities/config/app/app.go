package app

type App struct {
	Name string `yaml:"name" env-required:"true"`
}
