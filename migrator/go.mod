module github.com/kxddry/lectura/migrator

go 1.24.4

replace github.com/kxddry/lectura/shared => ../shared

require (
	github.com/golang-migrate/migrate/v4 v4.18.3
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/kxddry/lectura/shared v0.0.0-00010101000000-000000000000
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)
