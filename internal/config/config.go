package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	App      AppConfig      `yaml:"app"`
	KYC      KYCConfig      `yaml:"kyc"`
	Limits   LimitsConfig   `yaml:"limits"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type AppConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	JWTSecret  string `yaml:"jwt_secret"`
	HMACSecret string `yaml:"hmac_secret"`
}

type KYCConfig struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
}

type LimitsConfig struct {
	Tier1DailyLimit int64 `yaml:"tier1_daily_limit"`
	Tier2DailyLimit int64 `yaml:"tier2_daily_limit"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
