package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server           ServerConfig
	Database         DatabaseConfig
	Redis            RedisConfig
	JWT              JWTConfig
	SMS              SMSConfig
	Timeout          TimeoutConfig
	Security         SecurityConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Database        string
	Charset         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
	Expire time.Duration
}

type SMSConfig struct {
	Provider string
	Expire   time.Duration
}

type TimeoutConfig struct {
	Normal time.Duration
	Severe time.Duration
}

type SecurityConfig struct {
	EnableCodeVerification bool
	AllowedOrigins         []string
}

var GlobalConfig *Config

func Init(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
