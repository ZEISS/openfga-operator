package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config ...
type Config struct {
	OpenFGAURL string `envconfig:"OPENFGA_URL" default:"http://host.docker.internal:8080"`
}

// New ...
func New() *Config {
	return &Config{}
}

// Marshal ...
func (c *Config) Marshal() error {
	err := envconfig.Process("", c)
	if err != nil {
		return err
	}

	return nil
}
