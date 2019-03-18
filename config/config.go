package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	EnvironmentLocal = "local"
	EnvironmentTest  = "test"
)

type Config struct {
	Debug bool `yaml:"debug"`

	Environment string `yaml:"environment"`

	Server struct {
		Address string `yaml:"address"`
	} `yaml:"server"`

	DB struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		User         string `yaml:"user"`
		Password     string `yaml:"password"`
		Name         string `yaml:"name"`
		SSLMode      string `yaml:"ssl_mode"`
		MaxIdleConns int    `yaml:"max_idle_conns"`
		MaxOpenConns int    `yaml:"max_open_conns"`
	} `yaml:"db"`

	JWT struct {
		SecretKey     string `yaml:"secret_key"`
		ExpirySeconds int    `yaml:"expiry_seconds"`
	}

	Gnatsd struct {
		Log   bool `yaml:"log"`
		Debug bool `yaml:"debug"`
		Trace bool `yaml:"trace"`
		URL string `yaml:"url"`
	} `yaml:"gnatsd"`
}

func Parse(path string) (*Config, error) {
	cfg := Config{}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, err
}

func GetTestConfig() *Config {
	return &Config{
		Debug:       false,
		Environment: EnvironmentTest,
		Server: struct {
			Address string `yaml:"address"`
		}{
			Address: "localhost:5000",
		},
		DB: struct {
			Host         string `yaml:"host"`
			Port         int    `yaml:"port"`
			User         string `yaml:"user"`
			Password     string `yaml:"password"`
			Name         string `yaml:"name"`
			SSLMode      string `yaml:"ssl_mode"`
			MaxIdleConns int    `yaml:"max_idle_conns"`
			MaxOpenConns int    `yaml:"max_open_conns"`
		}{
			Host:         "localhost",
			Port:         5432,
			User:         "ben",
			Password:     "password",
			Name:         "chat_test_db",
			SSLMode:      "disable",
			MaxIdleConns: 2,
			MaxOpenConns: 20,
		},
		JWT: struct {
			SecretKey     string `yaml:"secret_key"`
			ExpirySeconds int    `yaml:"expiry_seconds"`
		}{
			SecretKey:     "replace_me",
			ExpirySeconds: 9999999,
		},
		Gnatsd: struct {
			Log   bool `yaml:"log"`
			Debug bool `yaml:"debug"`
			Trace bool `yaml:"trace"`
			URL   string `yaml:"url"`
		}{
			Log:   false,
			Debug: false,
			Trace: false,
			URL: "nats://localhost:4222",
		},
	}
}
