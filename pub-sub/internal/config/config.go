package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/zoninnik89/messenger/common"
	"os"
	"time"
)

type Config struct {
	Env   string      `yaml:"env" env-default:"local"`
	GRPC  GRPCConfig  `yaml:"grpc"`
	kafka KafkaConfig `yaml:"kafka"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type KafkaConfig struct {
	Port          int    `yaml:"port"`
	ConsumerID    string `yaml:"consumer_id"`
	ConsumerGroup string `yaml:"consumer_group"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config file path is empty")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist" + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())

	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "config file")
	flag.Parse()

	if res == "" {
		res = common.EnvString("CONFIG_PATH", "./config/local.yaml")
	}
	return res
}
