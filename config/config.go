package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port               string        `yaml:"port"`
	JWTSecret          string        `yaml:"jwt_secret"`
	KafkaBrokerAddress string        `yaml:"kafka_broker_address"`
	ReqPerSec          int           `yaml:"req_per_sec"`
	KafkaTopic         string        `yaml:"kafka_topic"`
	RateLimitWindow    time.Duration `yaml:"-"`
}

var cfg *Config

func GetConfig() *Config {
	return cfg
}

func LoadConfig(fileLoc string) error {
	var cfgVar Config
	f, err := os.ReadFile(fileLoc)
	if err != nil {
		return err
	}
	fenv := os.ExpandEnv(string(f))

	err = yaml.Unmarshal([]byte(fenv), &cfgVar)
	if err != nil {
		return err
	}
	if cfgVar.JWTSecret == "" {
		return fmt.Errorf("jwt_secret is required")
	}
	if cfgVar.KafkaBrokerAddress == "" {
		return fmt.Errorf("kafka_broker_address is required")
	}

	LoadDefault(&cfgVar)
	cfg = &cfgVar
	return nil
}

func LoadDefault(cfg *Config) {
	cfg.RateLimitWindow = time.Second
}
