package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

const (
	EnvironmentLocal = "local"
)

type Config struct {
	Debug bool `yaml:"debug"`

	Environment string `yaml:"environment"`

	Server struct {
		Address string `yaml:"address"`
	} `yaml:"server"`

	DB struct {
		Host string `yaml:"host"`
		Port int `yaml:"port"`
		User string `yaml:"user"`
		Password string `yaml:"password"`
		Name string `yaml:"name"`
		SSLMode string `yaml:"ssl_mode"`
	} `yaml:"db"`

	JWT struct {
		SecretKey string `yaml:"secret_key"`
		ExpirySeconds int `yaml:"expiry_seconds"`
	}

}

func Parse(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := Config{}
	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, err
}
