package config

import (
	"io/fs"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

func LoadConfig() (*Config, error) {
	var config Config

	// Load from env vars only if .env.local is found
	err := godotenv.Load(".env.local")
	if err != nil {
		if _, ok := err.(*fs.PathError); !ok {
			return nil, err
		}
	}

	if err := envconfig.Process("app", &config); err != nil {
		return nil, err
	}

	return &config, nil
}
