package main

import "github.com/savageking-io/ogbrest/kafka"

var (
	AppVersion     = "Undefined"
	ConfigFilepath = "rest-config.yaml"
	LogLevel       = ""
	AppConfig      Config
)

// Configuration structures for rest-config.yaml
type RestConfig struct {
	Hostname       string   `yaml:"hostname"`
	Port           uint16   `yaml:"port"`
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type ServiceConfig struct {
	Label    string `yaml:"label"`
	Hostname string `yaml:"hostname"`
	Port     uint16 `yaml:"port"`
	Token    string `yaml:"token"`
}

type UserClientConfig struct {
	Hostname string `yaml:"hostname"`
	Port     uint16 `yaml:"port"`
}

type Config struct {
	LogLevel   string           `yaml:"log_level"`
	Rest       RestConfig       `yaml:"rest"`
	Services   []ServiceConfig  `yaml:"services"`
	UserClient UserClientConfig `yaml:"user_client"`
	Kafka      kafka.Config     `yaml:"kafka"`
}
