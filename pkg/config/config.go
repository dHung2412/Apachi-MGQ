package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	ServerPort  string
	Environment string
	LogLevel    string

	// PostgreSQL
	DatabaseURL  string
	DatabaseHost string
	DatabasePort int
	DatabaseUser string
	DatabasePass string
	DatabaseName string

	// Memgraph (Bolt protocol)
	MemgraphURI  string
	MemgraphUser string
	MemgraphPass string

	// JWT Authentication
	JWTSecret string

	// Worker Pool
	WorkerCount    int
	ChannelBuffer  int
}

// Load reads configuration from .env files and environment variables.
// Environment variables take precedence over .env file values.
func Load(paths ...string) (*Config, error) {
	// Load .env file(s) if provided; ignore error if file doesn't exist
	if len(paths) == 0 {
		_ = godotenv.Load(".env")
	} else {
		for _, path := range paths {
			if err := godotenv.Load(path); err != nil {
				return nil, err
			}
		}
	}

	cfg := &Config{
		// Server
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),

		// PostgreSQL
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		DatabaseHost: getEnv("DB_HOST", "localhost"),
		DatabasePort: getEnvAsInt("DB_PORT", 5432),
		DatabaseUser: getEnv("DB_USER", "postgres"),
		DatabasePass: getEnv("DB_PASSWORD", ""),
		DatabaseName: getEnv("DB_NAME", "dp_maintenance"),

		// Memgraph
		MemgraphURI:  getEnv("MEMGRAPH_URI", "bolt://localhost:7687"),
		MemgraphUser: getEnv("MEMGRAPH_USER", ""),
		MemgraphPass: getEnv("MEMGRAPH_PASS", ""),

		// JWT
		JWTSecret: getEnv("JWT_SECRET", ""),

		// Worker Pool
		WorkerCount:   getEnvAsInt("WORKER_COUNT", 5),
		ChannelBuffer: getEnvAsInt("CHANNEL_BUFFER", 1000),
	}

	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET environment variable is required")
	}

	return cfg, nil
}

// GetDSN builds a PostgreSQL connection string from individual fields.
func (c *Config) GetDSN() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	return "host=" + c.DatabaseHost +
		" port=" + strconv.Itoa(c.DatabasePort) +
		" user=" + c.DatabaseUser +
		" password=" + c.DatabasePass +
		" dbname=" + c.DatabaseName +
		" sslmode=disable"
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
