package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	App AppConfig
	DB  DBConfig
	Log LogConfig
	JWT JWTConfig
}

type JWTConfig struct {
	Secret string
	TTL    time.Duration
}

type AppConfig struct {
	Env  string
	Port string
}

type DBConfig struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func (databaseConfig DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		databaseConfig.Host, databaseConfig.Port, databaseConfig.Name, databaseConfig.User, databaseConfig.Password, databaseConfig.SSLMode,
	)
}

type LogConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	maxOpenConnections, err := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %w", err)
	}

	maxIdleConnections, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE_CONNS: %w", err)
	}

	connectionMaxLifetime, err := time.ParseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME: %w", err)
	}

	jwtTTL, err := time.ParseDuration(getEnv("JWT_TTL", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_TTL: %w", err)
	}

	configuration := &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "change-me-in-production"),
			TTL:    jwtTTL,
		},
		DB: DBConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			Name:            getEnv("DB_NAME", "fintech_db"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    maxOpenConnections,
			MaxIdleConns:    maxIdleConnections,
			ConnMaxLifetime: connectionMaxLifetime,
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	return configuration, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
