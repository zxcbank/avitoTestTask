package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Port                    string        `yaml:"port" env-required:"true"`
	Timeout                 time.Duration `yaml:"timeout" env-default:"300ms"`
	IdleTimeout             time.Duration `yaml:"idle_timeout" env-default:"60s"`
	GracefulShutdownTimeOut time.Duration `yaml:"graceful_shutdown_time_out" env-default:"3600s"`
}

func MustLoad() *Config {
	os.Setenv("CONFIG_PATH", "config/local.yaml")
	config := os.Getenv("CONFIG_PATH")
	if config == "" {
		log.Fatal("CONFIG_PATH environment variable not set")
	}
	if _, err := os.Stat(config); os.IsNotExist(err) {
		log.Fatal("CONFIG_PATH does not exist")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(config, &cfg); err != nil {
		log.Fatal("can't read config file: ", err)
	}

	return &cfg
}
