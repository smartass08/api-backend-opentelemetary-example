package config

import (
	"github.com/spf13/viper"
)

var (
	cfg *Config
)

type Config struct {
	Port         string
	Environment  string
	LogLevel     string
	OTLPEndpoint string
	OtelAPIKey   string
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("PORT", "3000")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("OTLP_ENDPOINT", "")
	viper.SetDefault("OTEL_API_KEY", "")

	cfg = &Config{
		Port:         viper.GetString("PORT"),
		Environment:  viper.GetString("ENV"),
		LogLevel:     viper.GetString("LOG_LEVEL"),
		OTLPEndpoint: viper.GetString("OTLP_ENDPOINT"),
		OtelAPIKey:   viper.GetString("OTEL_API_KEY"),
	}

	return cfg
}

func GetConfig() *Config {
	if cfg == nil {
		LoadConfig()
	}
	return cfg
}
