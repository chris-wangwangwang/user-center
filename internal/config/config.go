package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort int

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret    string
	JWTIssuer    string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

var instance *Config
var once bool

func Get() *Config {
	if !once {
		instance = load()
		once = true
	}
	return instance
}

func load() *Config {
	return &Config{
		HTTPPort: getEnvInt("HTTP_PORT", 8080),

		DBHost:     getEnv("DB_HOST", "172.20.90.73"),
		DBPort:     getEnvInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "123456"),
		DBName:     getEnv("DB_NAME", "usercenter"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production-please"),
		JWTIssuer:      getEnv("JWT_ISSUER", "usercenter"),
		AccessTokenTTL: time.Duration(getEnvInt("ACCESS_TOKEN_TTL_MIN", 60)) * time.Minute,
		RefreshTokenTTL: time.Duration(getEnvInt("REFRESH_TOKEN_TTL_HOUR", 24*7)) * time.Hour,
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
