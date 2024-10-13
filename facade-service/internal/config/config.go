package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env        string `yaml:"env" env:"ENV" env-default:"local"`
	HTTPServer `yaml:"http_server"`
	Consul     ConsulConfig `yaml:"consul"`
}

type HTTPServer struct {
	Port        int           `yaml:"port" env:"PORT" env-default:"8080"`
	Address     string        `yaml:"address" env:"ADDRESS" env-default:"localhost"`
	Timeout     time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idleTimeout" env-default:"10s"`
	Name        string        `yaml:"name" env:"NAME" env-default:"facade"`
}

type ConsulConfig struct {
	Port int `yaml:"port"`
}

func MustLoad() *Config {
	config := os.Getenv("CONFIG_PATH")
	if config == "" {
		log.Fatal("CONFIG_PATH environment variable not set")
	}

	if _, err := os.Stat(config); os.IsNotExist(err) {
		log.Fatal("CONFIG_PATH does not exist")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(config, &cfg); err != nil {
		log.Fatalf("can't read config file: %s", err)
	}

	return &cfg
}
