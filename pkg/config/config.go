package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Addr     string `yaml:"addr"`
	TLS      bool   `yaml:"tls"`
	CertFile string `yaml:"cert_file"` // Optional: Path to cert. If empty, auto-generate.
	KeyFile  string `yaml:"key_file"`  // Optional: Path to key.

	EndpointsPath string `yaml:"endpoints_path"` // Optional: Path that lists configured routes. Defaults to /endpoints.
}

type Route struct {
	Path    string            `yaml:"path"`
	Target  string            `yaml:"target"`
	Headers map[string]string `yaml:"headers"` // Generic headers to inject (e.g. Authorization)
}

type Config struct {
	Server ServerConfig `yaml:"server"`
	Routes []Route      `yaml:"routes"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
