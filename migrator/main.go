package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/kxddry/lectura/shared/entities/config/db"
	"os"
	"strings"

	// migration
	"github.com/golang-migrate/migrate/v4"

	// drivers
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type MigrationConfig struct {
	St            db.StorageConfig `yaml:"storage" env-required:"true"` // use dbname = postgres here
	Operation     string           `env:"OPERATION" yaml:"operation" env-default:"up"`
	DbsMigrations []Entry          `yaml:"dbs_migrations" env-required:"true"`
}

type Entry struct {
	Name string `yaml:"name"` // DBName
	Path string `yaml:"path"` // Path for migrations
}

func main() {
	confPath := os.Getenv("CONFIG_PATH")
	if confPath == "" {
		panic("CONFIG_PATH env variable not set")
	}

	var cfg MigrationConfig
	if err := cleanenv.ReadConfig(confPath, &cfg); err != nil {
		panic(err)
	}
	op := cfg.Operation

	if len(cfg.DbsMigrations) == 0 {
		panic("No dbs migrations configured")
	}

	for _, m := range cfg.DbsMigrations {
		name, path := m.Name, m.Path
		ccfg := cfg.St
		ccfg.DBName = name
		err := EnsureDBexists(name, ccfg)
		if err != nil {
			panic(err)
		}
		link := Link(ccfg)

		DoOneMigration(link, path, op)
	}

	fmt.Println("migration successful")
}

func DoOneMigration(link, path, op string) {
	m, err := migrate.New("file://"+path, link)
	if err != nil {
		panic(err)
	}
	switch {
	case op == "" || op == "up":
		if err = m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("Nothing to migrate at", path)
				return
			}
			panic(err)
		}
		return
	case op == "down":
		if err = m.Force(1); err != nil {
			panic(err)
		}
		if err = m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("Nothing to migrate at", path)
				return
			}
			panic(err)
		}
	default:
		panic("Unknown operation: " + op)
	}
}

func Link(cfg db.StorageConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
}

func DataSourceName(cfg db.StorageConfig) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
}

func EnsureDBexists(dbname string, adminCfg db.StorageConfig) error {
	adminCfg.DBName = "postgres"
	_db, err := sql.Open("postgres", DataSourceName(adminCfg))
	if err != nil {
		return err
	}
	defer _db.Close()

	_, err = _db.Exec("CREATE DATABASE " + dbname)
	if err != nil && !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "no change") {
		return err
	}
	return nil
}
