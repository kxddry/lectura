package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/kxddry/lectura/shared/entities/config/db"
	cc "github.com/kxddry/lectura/updater/internal/config"
	"log"
	"os"
	"strings"

	// migration
	"github.com/golang-migrate/migrate/v4"

	// drivers
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// USAGE:
// --config=/path/to/config.yaml
// inside config.yaml:
// Storage: host, port, user, password, dbname, sslmode
// Migrations: /path/to/migrations/*.sql - folder
func main() {
	var op string
	flag.StringVar(&op, "operation", "", "operation: up or down")
	if op == "" {
		op = os.Getenv("OPERATION")
	}
	confPath := os.Getenv("CONFIG_PATH")
	if confPath == "" {
		panic("CONFIG_PATH env variable not set")
	}

	var cfg cc.Config
	if err := cleanenv.ReadConfig(confPath, &cfg); err != nil {
		panic(err)
	}

	pSt := cfg.Storage
	pSt.DBName = "postgres"
	dsn := DataSourceName(pSt)
	link := Link(cfg.Storage)
	err := EnsureDBexists(cfg.Storage.DBName, dsn)
	if err != nil {
		panic(err)
	}

	m, err := migrate.New("file://"+cfg.Migrations.Path, link)
	if err != nil {
		panic(err)
	}
	switch {
	case op == "" || op == "up":
		if err = m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("Nothing to migrate")
				return
			}
			panic(err)
		}
	case op == "down":
		if err = m.Force(1); err != nil {
			panic(err)
		}
		if err = m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("Nothing to migrate")
				return
			}
			panic(err)
		}
	default:
		log.Fatalln("Unknown operation:", op)
	}

	log.Println("migration successful")
}

func Link(cfg db.StorageConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
}

func DataSourceName(cfg db.StorageConfig) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
}

func EnsureDBexists(dbname, dsn string) error {
	_db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = _db.Close() }()

	_, err = _db.Exec("CREATE DATABASE" + " " + dbname)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}
	return nil
}
