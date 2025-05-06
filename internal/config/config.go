package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Env         string `yaml:"env" env-default:"prod"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:5050"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func LoadConfig() (config Configuration, err error) {
	var data []byte
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH not found")
	}

	data, err = os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Read err")
	}
	err = yaml.Unmarshal(data, &config)
	return
}
