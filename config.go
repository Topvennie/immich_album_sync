package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ImmichURL    string       `yaml:"immich_url"`
	PollInterval int          `yaml:"poll_interval"`
	Users        []UserConfig `yaml:"users"`
}

type UserConfig struct {
	APIKey string   `yaml:"api_key"`
	Paths  []string `yaml:"paths"`
}

func loadConfig() Config {
	file, err := os.ReadFile("config.yml")
	if err != nil {
		log.Fatal(err)
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		log.Fatal(err)
	}

	if len(config.ImmichURL) > 0 && config.ImmichURL[len(config.ImmichURL)-1] != '/' {
		config.ImmichURL = config.ImmichURL + "/"
	}

	return config
}
