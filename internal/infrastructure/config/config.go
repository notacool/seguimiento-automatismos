package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// Load carga la configuración desde variables de entorno
func Load() (*Config, error) {
	dbPort, err := strconv.Atoi(getEnv("DATABASE_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_PORT: %w", err)
	}

	maxConns, err := strconv.Atoi(getEnv("DATABASE_MAX_CONNS", "25"))
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_MAX_CONNS: %w", err)
	}

	minConns, err := strconv.Atoi(getEnv("DATABASE_MIN_CONNS", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_MIN_CONNS: %w", err)
	}

	maxConnLifetime, err := time.ParseDuration(getEnv("DATABASE_MAX_CONN_LIFETIME", "5m"))
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_MAX_CONN_LIFETIME: %w", err)
	}

	maxConnIdleTime, err := time.ParseDuration(getEnv("DATABASE_MAX_CONN_IDLE_TIME", "1m"))
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_MAX_CONN_IDLE_TIME: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DATABASE_HOST", "localhost"),
			Port:            dbPort,
			User:            getEnv("DATABASE_USER", "postgres"),
			Password:        getEnv("DATABASE_PASSWORD", "postgres"),
			Name:            getEnv("DATABASE_NAME", "proceslog"),
			SSLMode:         getEnv("DATABASE_SSLMODE", "disable"),
			MaxConns:        maxConns,
			MinConns:        minConns,
			MaxConnLifetime: maxConnLifetime,
			MaxConnIdleTime: maxConnIdleTime,
		},
	}, nil
}

// ConnectionString genera la cadena de conexión a PostgreSQL
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
