package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	HTTPPort int `toml:"http_port"`

	DBHost     string `toml:"db_host"`
	DBPort     int    `toml:"db_port"`
	DBUser     string `toml:"db_user"`
	DBPassword string `toml:"db_password"`
	DBName     string `toml:"db_name"`
	DBSSLMode  string `toml:"db_sslmode"`

	JWTSecret             string `toml:"jwt_secret"`
	JWTIssuer             string `toml:"jwt_issuer"`
	AccessTokenTTLMinutes int    `toml:"access_token_ttl_minutes"`
	RefreshTokenTTLHours  int    `toml:"refresh_token_ttl_hours"`
}

func (c *Config) AccessTokenTTL() time.Duration {
	return time.Duration(c.AccessTokenTTLMinutes) * time.Minute
}

func (c *Config) RefreshTokenTTL() time.Duration {
	return time.Duration(c.RefreshTokenTTLHours) * time.Hour
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

var (
	instance *Config
	loaded   bool
)

func Get() *Config {
	if !loaded {
		instance = load()
		loaded = true
	}
	return instance
}

const configPath = "config.toml"

func defaults() *Config {
	return &Config{
		HTTPPort: 8080,

		DBHost:     "172.20.90.73",
		DBPort:     5432,
		DBUser:     "postgres",
		DBPassword: "123456",
		DBName:     "usercenter",
		DBSSLMode:  "disable",

		JWTSecret:             "change-me-in-production-please",
		JWTIssuer:             "usercenter",
		AccessTokenTTLMinutes: 60,
		RefreshTokenTTLHours:  24 * 7,
	}
}

func load() *Config {
	cfg := defaults()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("config: read %s failed, fall back to defaults: %v", configPath, err)
		} else {
			log.Printf("config: %s not found, using built-in defaults", configPath)
		}
		return cfg
	}
	if err := toml.Unmarshal(data, cfg); err != nil {
		log.Fatalf("config: parse %s failed: %v", configPath, err)
	}
	return cfg
}
