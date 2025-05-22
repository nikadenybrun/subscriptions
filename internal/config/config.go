package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `env:"env" env-default:"local"`
	StoragePath string `env:"STORAGE_PATH"`
	HTTPServer
	DBSaver
	AppPort string `env:"APP_PORT"`
}

type HTTPServer struct {
	Address     string        `env:"HTTP_ADDRESS" env-default:"localhost:8080"`
	Timeout     time.Duration `env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"30s"`
	User        string        `env:"HTTP_USER" env-required:"true"`
	Password    string        `env:"HTTP_PASSWORD" env-required:"true"`
}

type DBSaver struct {
	DbUser string `env:"POSTGRES_USER"`
	DbPass string `env:"POSTGRES_PASSWORD"`
	DbHost string `env:"POSTGRES_HOST"`
	DbAddr string `env:"POSTGRES_ADDRESS"`
	DbName string `env:"POSTGRES_DB"`
}

func MustLoad() *Config {

	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	fmt.Println(cfg)

	return &cfg
}
